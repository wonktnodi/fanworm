package demo

import (
    "os"
    "net"
    "syscall"
    "time"
    "sync"
    "sort"
    "github.com/wonktnodi/go-revolver/pkg/base"
)

type PollingReactor struct {
    PollBase
    fdconn       map[int]*connection
    idconn       map[int]*connection
    lns          []*listener
    stdlib       bool
    timeoutqueue *TimeoutQueue
    mu           sync.Mutex
    done         bool
}

func NewReactor() (inst *PollingReactor) {
    r := PollingReactor{}
    inst = &r
    return
}

func (r *PollingReactor) parseAddr(addr ...string) (err error) {
    var stdlib bool
    for _, addr := range addr {
        var ln listener
        var stdlibt bool
        ln.network, ln.addr, stdlibt = parseAddr(addr)
        if stdlibt {
            r.stdlib = true
        }
        if ln.network == "unix" {
            os.RemoveAll(ln.addr)
        }

        ln.ln, err = net.Listen(ln.network, ln.addr)
        if err != nil {
            return err
        }
        ln.naddr = ln.ln.Addr()
        if !stdlib {
            if err := ln.system(); err != nil {
                return err
            }
        }
        ln.reactor = r
        r.lns = append(r.lns, &ln)
    }

    // add event accept event monitoring
    for _, ln := range r.lns {
        if err := r.registerEventHandler(ln, EventRead); err != nil {
            return err
        }
    }
    return
}

func (r *PollingReactor) Open() (err error) {
    p, err := base.MakePoll()
    if err != nil {
        return err
    }
    r.p = p
    makeEvents(&r.PollBase, 64)

    r.fdconn = make(map[int]*connection)
    r.idconn = make(map[int]*connection)

    r.timeoutqueue = NewTimeoutQueue()
    return
}

func (r *PollingReactor) Close() {
    syscall.Close(r.p)

    for _, ln := range r.lns {
        ln.close()
    }
}

func (r *PollingReactor) Serve(events Events, addrs ...string) (err error) {
    defer func() {
        r.Close()
    }()

    r.parseAddr(addrs...)
    return r.serve(events, r.lns)
}

func (r *PollingReactor) lock() {
    r.mu.Lock()
}

func (r *PollingReactor) unlock() {
    r.mu.Unlock()
}

func (r *PollingReactor) serve(events Events, lns []*listener) error {
    // wake wakes up a connection
    ctx := Server{Wake: r.wake, Dial: r.dail}
    ctx.Addrs = make([]net.Addr, len(lns))
    for i, ln := range lns {
        ctx.Addrs[i] = ln.naddr
    }
    if events.Serving != nil {
        switch events.Serving(ctx) {
        case Shutdown:
            return nil
        }
    }
    // cleaning work
    defer r.clean(events)

    //var rsa syscall.Sockaddr

    var packet [0xFFFF]byte
    nextTicker := time.Now()
    // event looping
    for {
        delay := nextTicker.Sub(time.Now())
        if delay < 0 {
            delay = 0
        } else if delay > time.Second/4 {
            delay = time.Second / 4
        }
        pn, err := base.Wait(r.p, r.events, delay)
        if err != nil && err != syscall.EINTR {
            return err
        }
        remain := nextTicker.Sub(time.Now())
        if remain < 0 {
            var tickerDelay time.Duration
            var action Action
            if events.Tick != nil {
                tickerDelay, action = events.Tick()
                if action == Shutdown {
                    return nil
                }
            } else {
                tickerDelay = time.Hour
            }
            nextTicker = time.Now().Add(tickerDelay + remain)
        }
        // check for dial connection timeouts
        if r.timeoutqueue.Len() > 0 {
            var count int
            now := time.Now()
            for {
                v := r.timeoutqueue.Peek()
                if v == nil {
                    break
                }
                c := v.(*connection)
                if now.After(v.Timeout()) {
                    r.timeoutqueue.Pop()
                    if _, ok := connMgr.idconn[c.id]; ok && c.opening {
                        delete(r.idconn, c.id)
                        delete(r.fdconn, c.fd)
                        connMgr.RemoveConnection(c)
                        genAddrs(c)
                        syscall.Close(c.fd)
                        if events.Opened != nil {
                            events.Opened(c.id, Info{
                                Closing:    true,
                                AddrIndex:  c.lnidx,
                                LocalAddr:  c.laddr,
                                RemoteAddr: c.raddr,
                            })
                        }
                        if events.Closed != nil {
                            if c.dialerr != nil {
                                events.Closed(c.id, c.dialerr)
                            } else {
                                events.Closed(c.id, syscall.ETIMEDOUT)
                            }
                        }
                        count++
                    }
                } else {
                    break
                }
            }
            if count > 0 {
                // invalidate the current events and wait for more
                continue
            }
        }
        r.lock()
        for i := 0; i < pn; i++ {
            var in []byte
            var c *connection
            var nfd int
            var n int
            var out []byte
            var ln *listener
            var lnidx int
            var fd = getFD(&r.PollBase, i)

            for lnidx, ln = range lns {
                if fd == ln.fd {
                    c, err = ln.Accept()
                    if err != nil {
                        goto fail
                    }

                    r.fdconn[nfd] = c
                    r.idconn[c.id] = c
                    goto next
                    //goto accept
                }
            }
            ln = nil
            c = r.fdconn[fd]
            c = connMgr.GetConnection(fd)
            if c == nil {
                syscall.Close(fd)
                goto next
            }
            if c.opening {
                goto opened
            }
            goto read
        opened:
            genAddrs(c)
            if err = base.AddRead(r.p, c.fd, &c.readon, &c.writeon); err != nil {
                goto fail
            }
            if events.Opened != nil {
                r.unlock()
                out, c.opts, c.action = events.Opened(c.id, Info{
                    AddrIndex:  lnidx,
                    LocalAddr:  c.laddr,
                    RemoteAddr: c.raddr,
                })
                r.lock()
                if c.opts.TCPKeepAlive > 0 {
                    base.SetKeepAlive(c.fd, int(c.opts.TCPKeepAlive/time.Second))
                }
                if len(out) > 0 {
                    c.outbuf = append(c.outbuf, out...)
                }
            }
            if c.opening {
                c.opening = false
                goto next
            }
            goto write
        read:
            if c.action != None {
                goto write
            }
            if c.wake {
                c.wake = false
            } else {
                n, err = c.Read(packet[:])
                if n == 0 || err != nil {
                    if err == syscall.EAGAIN {
                        goto write
                    }
                    c.err = err
                    goto close
                }
                in = append([]byte{}, packet[:n]...)
            }
            // if c.laddr == nil {
            // 	// we need the local address and to open the socket
            // 	lsa, _ = syscall.Getsockname(c.fd)
            // 	c.laddr = sock
            // }
            if events.Data != nil {
                r.unlock()
                out, c.action = events.Data(c.id, in)
                r.lock()
            }
            if len(out) > 0 {
                c.outbuf = append(c.outbuf, out...)
            }
            goto write
        write:
            if len(c.outbuf)-c.outpos > 0 {
                if events.Prewrite != nil {
                    r.unlock()
                    action := events.Prewrite(c.id, len(c.outbuf[c.outpos:]))
                    r.lock()
                    if action == Shutdown {
                        c.action = Shutdown
                    }
                }
                n, err = c.Write(c.outbuf[c.outpos:])
                if events.Postwrite != nil {
                    amount := n
                    if amount < 0 {
                        amount = 0
                    }
                    r.unlock()
                    action := events.Postwrite(c.id, amount, len(c.outbuf)-c.outpos-amount)
                    r.lock()
                    if action == Shutdown {
                        c.action = Shutdown
                    }
                }
                if n == 0 || err != nil {
                    if c.action == Shutdown {
                        goto close
                    }
                    if err == syscall.EAGAIN {
                        if err = base.AddWrite(r.p, c.fd, &c.readon, &c.writeon); err != nil {
                            goto fail
                        }
                        goto next
                    }
                    c.err = err
                    goto close
                }
                c.outpos += n
                if len(c.outbuf)-c.outpos == 0 {
                    c.outpos = 0
                    c.outbuf = c.outbuf[:0]
                }
            }
            if c.action == Shutdown {
                goto close
            }
            if len(c.outbuf)-c.outpos == 0 {
                if !c.wake {
                    if err = base.DelWrite(r.p, c.fd, &c.readon, &c.writeon); err != nil {
                        goto fail
                    }
                }
                if c.action != None {
                    goto close
                }
            } else {
                if err = base.AddWrite(r.p, c.fd, &c.readon, &c.writeon); err != nil {
                    goto fail
                }
            }
            goto next
        close:
            delete(r.fdconn, c.fd)
            delete(r.idconn, c.id)
            connMgr.RemoveConnection(c)
            if c.action == Detach {
                if events.Detached != nil {
                    c.detached = true
                    if len(c.outbuf)-c.outpos > 0 {
                        c.outbuf = append(c.outbuf[:0], c.outbuf[c.outpos:]...)
                    } else {
                        c.outbuf = nil
                    }
                    c.outpos = 0
                    syscall.SetNonblock(c.fd, false)
                    r.unlock()
                    c.action = events.Detached(c.id, c)
                    r.lock()
                    if c.action == Shutdown {
                        goto fail
                    }
                    goto next
                }
            }
            syscall.Close(c.fd)
            if events.Closed != nil {
                r.unlock()
                action := events.Closed(c.id, c.err)
                r.lock()
                if action == Shutdown {
                    c.action = Shutdown
                }
            }
            if c.action == Shutdown {
                err = nil
                goto fail
            }
            goto next
        fail:
            r.unlock()
            return err
        next:
        }
        r.unlock()
    }
}

func (r *PollingReactor) dail(addr string, timeout time.Duration) int {
    r.lock()
    if r.done {
        r.unlock()
        return 0
    }
    c := &connection{id: connMgr.GetID(), opening: true, lnidx: -1}
    r.idconn[c.id] = c
    connMgr.AddConnection(c)
    if timeout != 0 {
        c.timeout = time.Now().Add(timeout)
        r.timeoutqueue.Push(c)
    }
    r.unlock()
    // resolving an address blocks and we don't want blocking, like ever.
    // but since we're leaving the event loop we'll need to complete the
    // socket connection in a goroutine and add the read and write events
    // to the loop to get back into the loop.
    go func() {
        err := func() error {
            sa, err := resolve(addr)
            if err != nil {
                return err
            }
            var fd int
            switch sa.(type) {
            case *syscall.SockaddrUnix:
                fd, err = syscall.Socket(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
            case *syscall.SockaddrInet4:
                fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
            case *syscall.SockaddrInet6:
                fd, err = syscall.Socket(syscall.AF_INET6, syscall.SOCK_STREAM, 0)
            }
            if err != nil {
                return err
            }
            err = syscall.Connect(fd, sa)
            if err != nil && err != syscall.EINPROGRESS {
                syscall.Close(fd)
                return err
            }
            if err := syscall.SetNonblock(fd, true); err != nil {
                syscall.Close(fd)
                return err
            }
            r.lock()
            err = base.AddRead(r.p, fd, &c.readon, &c.writeon)
            if err != nil {
                r.unlock()
                syscall.Close(fd)
                return err
            }
            err = base.AddWrite(r.p, fd, &c.readon, &c.writeon)
            if err != nil {
                r.unlock()
                syscall.Close(fd)
                return err
            }
            c.fd = fd
            r.fdconn[fd] = c
            connMgr.AddConnection(c)
            r.unlock()
            return nil
        }()
        if err != nil {
            // set a dial error and timeout right away
            r.lock()
            c.dialerr = err
            c.timeout = time.Now()
            r.timeoutqueue.Push(c)
            r.unlock()
        }
    }()
    return c.id
}

func (r *PollingReactor) wake(id int) bool {
    var ok = true
    var err error
    r.lock()
    if r.done {
        r.unlock()
        return false
    }
    c := r.idconn[id]
    c = connMgr.GetConnectionByID(id)
    if c == nil || c.fd == 0 {
        ok = false
    } else if !c.wake {
        c.wake = true
        err = base.AddWrite(r.p, c.fd, &c.readon, &c.writeon)
    }
    r.unlock()
    if err != nil {
        panic(err)
    }
    return ok
}

func (r *PollingReactor) clean(events Events) {
    r.lock()
    r.done = true
    type fdid struct {
        fd, id  int
        opening bool
        laddr   net.Addr
        raddr   net.Addr
        lnidx   int
    }
    var fdids []fdid
    for _, c := range connMgr.idconn {
        if c.opening {
            genAddrs(c)
        }
        fdids = append(fdids, fdid{c.fd, c.id, c.opening, c.laddr, c.raddr, c.lnidx})
    }
    sort.Slice(fdids, func(i, j int) bool {
        return fdids[j].id < fdids[i].id
    })
    for _, fdid := range fdids {
        if fdid.fd != 0 {
            syscall.Close(fdid.fd)
        }
        if fdid.opening {
            if events.Opened != nil {
                r.unlock()
                events.Opened(fdid.id, Info{
                    Closing:    true,
                    AddrIndex:  fdid.lnidx,
                    LocalAddr:  fdid.laddr,
                    RemoteAddr: fdid.raddr,
                })
                r.lock()
            }
        }
        if events.Closed != nil {
            r.unlock()
            events.Closed(fdid.id, nil)
            r.lock()
        }
    }
    syscall.Close(r.p)
    r.fdconn = nil
    r.idconn = nil
    r.unlock()
}
