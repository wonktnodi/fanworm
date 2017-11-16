package base

import (
    "github.com/wonktnodi/go-revolver/internal"
    "syscall"
    "sync"
    "time"
    "net"
    "sort"
    "io"
    "sync/atomic"
)

func serve(events Events, lns []*listener) error {
    p, err := internal.MakePoll()
    if err != nil {
        return err
    }
    defer syscall.Close(p)
    for _, ln := range lns {
        if err := internal.AddRead(p, ln.fd, nil, nil); err != nil {
            return err
        }
    }
    var mu sync.Mutex
    var done bool
    lock := func() { mu.Lock() }
    unlock := func() { mu.Unlock() }
    fdconn := make(map[int]*unixConn)
    idconn := make(map[int]*unixConn)
    timeoutqueue := internal.NewTimeoutQueue()
    var id int
    dial := func(addr string, timeout time.Duration) int {
        lock()
        if done {
            unlock()
            return 0
        }
        id++
        c := &unixConn{id: id, opening: true, lnidx: -1}
        idconn[id] = c
        if timeout != 0 {
            c.timeout = time.Now().Add(timeout)
            timeoutqueue.Push(c)
        }
        unlock()
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
                lock()
                err = internal.AddRead(p, fd, &c.readon, &c.writeon)
                if err != nil {
                    unlock()
                    syscall.Close(fd)
                    return err
                }
                err = internal.AddWrite(p, fd, &c.readon, &c.writeon)
                if err != nil {
                    unlock()
                    syscall.Close(fd)
                    return err
                }
                c.fd = fd
                fdconn[fd] = c
                unlock()
                return nil
            }()
            if err != nil {
                // set a dial error and timeout right away
                lock()
                c.dialerr = err
                c.timeout = time.Now()
                timeoutqueue.Push(c)
                unlock()
            }

        }()
        return id
    }

    // wake wakes up a connection
    wake := func(id int) bool {
        var ok = true
        var err error
        lock()
        if done {
            unlock()
            return false
        }
        c := idconn[id]
        if c == nil || c.fd == 0 {
            ok = false
        } else if !c.wake {
            c.wake = true
            err = internal.AddWrite(p, c.fd, &c.readon, &c.writeon)
        }
        unlock()
        if err != nil {
            panic(err)
        }
        return ok
    }
    ctx := Server{Wake: wake, Dial: dial}
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
    defer func() {
        lock()
        done = true
        type fdid struct {
            fd, id  int
            opening bool
            laddr   net.Addr
            raddr   net.Addr
            lnidx   int
        }
        var fdids []fdid
        for _, c := range idconn {
            if c.opening {
                filladdrs(c)
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
                    unlock()
                    events.Opened(fdid.id, Info{
                        Closing:    true,
                        AddrIndex:  fdid.lnidx,
                        LocalAddr:  fdid.laddr,
                        RemoteAddr: fdid.raddr,
                    })
                    lock()
                }
            }
            if events.Closed != nil {
                unlock()
                events.Closed(fdid.id, nil)
                lock()
            }
        }
        syscall.Close(p)
        fdconn = nil
        idconn = nil
        unlock()
    }()
    var rsa syscall.Sockaddr

    var packet [0xFFFF]byte
    var evs = internal.MakeEvents(64)
    nextTicker := time.Now()
    for {
        delay := nextTicker.Sub(time.Now())
        if delay < 0 {
            delay = 0
        } else if delay > time.Second/4 {
            delay = time.Second / 4
        }
        pn, err := internal.Wait(p, evs, delay)
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
        if timeoutqueue.Len() > 0 {
            var count int
            now := time.Now()
            for {
                v := timeoutqueue.Peek()
                if v == nil {
                    break
                }
                c := v.(*unixConn)
                if now.After(v.Timeout()) {
                    timeoutqueue.Pop()
                    if _, ok := idconn[c.id]; ok && c.opening {
                        delete(idconn, c.id)
                        delete(fdconn, c.fd)
                        filladdrs(c)
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
        lock()
        for i := 0; i < pn; i++ {
            var in []byte
            var c *unixConn
            var nfd int
            var n int
            var out []byte
            var ln *listener
            var lnidx int
            var fd = internal.GetFD(evs, i)
            for lnidx, ln = range lns {
                if fd == ln.fd {
                    goto accept
                }
            }
            ln = nil
            c = fdconn[fd]
            if c == nil {
                syscall.Close(fd)
                goto next
            }
            if c.opening {
                goto opened
            }
            goto read
        accept:
            nfd, rsa, err = syscall.Accept(fd)
            if err != nil {
                goto next
            }
            if err = syscall.SetNonblock(nfd, true); err != nil {
                goto fail
            }
            id++
            c = &unixConn{id: id, fd: nfd,
                opening: true,
                lnidx:   lnidx,
                raddr:   sockaddrToAddr(rsa),
            }
            // we have a remote address but the local address yet.
            if err = internal.AddWrite(p, c.fd, &c.readon, &c.writeon); err != nil {
                goto fail
            }
            fdconn[nfd] = c
            idconn[id] = c
            goto next
        opened:
            filladdrs(c)
            if err = internal.AddRead(p, c.fd, &c.readon, &c.writeon); err != nil {
                goto fail
            }
            if events.Opened != nil {
                unlock()
                out, c.opts, c.action = events.Opened(c.id, Info{
                    AddrIndex:  lnidx,
                    LocalAddr:  c.laddr,
                    RemoteAddr: c.raddr,
                })
                lock()
                if c.opts.TCPKeepAlive > 0 {
                    internal.SetKeepAlive(c.fd, int(c.opts.TCPKeepAlive/time.Second))
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
                unlock()
                out, c.action = events.Data(c.id, in)
                lock()
            }
            if len(out) > 0 {
                c.outbuf = append(c.outbuf, out...)
            }
            goto write
        write:
            if len(c.outbuf)-c.outpos > 0 {
                if events.Prewrite != nil {
                    unlock()
                    action := events.Prewrite(c.id, len(c.outbuf[c.outpos:]))
                    lock()
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
                    unlock()
                    action := events.Postwrite(c.id, amount, len(c.outbuf)-c.outpos-amount)
                    lock()
                    if action == Shutdown {
                        c.action = Shutdown
                    }
                }
                if n == 0 || err != nil {
                    if c.action == Shutdown {
                        goto close
                    }
                    if err == syscall.EAGAIN {
                        if err = internal.AddWrite(p, c.fd, &c.readon, &c.writeon); err != nil {
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
                    if err = internal.DelWrite(p, c.fd, &c.readon, &c.writeon); err != nil {
                        goto fail
                    }
                }
                if c.action != None {
                    goto close
                }
            } else {
                if err = internal.AddWrite(p, c.fd, &c.readon, &c.writeon); err != nil {
                    goto fail
                }
            }
            goto next
        close:
            delete(fdconn, c.fd)
            delete(idconn, c.id)
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
                    unlock()
                    c.action = events.Detached(c.id, c)
                    lock()
                    if c.action == Shutdown {
                        goto fail
                    }
                    goto next
                }
            }
            syscall.Close(c.fd)
            if events.Closed != nil {
                unlock()
                action := events.Closed(c.id, c.err)
                lock()
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
            unlock()
            return err
        next:
        }
        unlock()
    }
}

// resolve resolves an evio address and retuns a sockaddr for socket
// connection to external servers.
func resolve(addr string) (sa syscall.Sockaddr, err error) {
    network, address, _ := parseAddr(addr)
    var taddr net.Addr
    switch network {
    default:
        return nil, net.UnknownNetworkError(network)
    case "unix":
        taddr = &net.UnixAddr{Net: "unix", Name: address}
    case "tcp", "tcp4", "tcp6":
        // use the stdlib resolver because it's good.
        taddr, err = net.ResolveTCPAddr(network, address)
        if err != nil {
            return nil, err
        }
    }
    switch taddr := taddr.(type) {
    case *net.UnixAddr:
        sa = &syscall.SockaddrUnix{Name: taddr.Name}
    case *net.TCPAddr:
        switch len(taddr.IP) {
        case 0:
            var sa4 syscall.SockaddrInet4
            sa4.Port = taddr.Port
            sa = &sa4
        case 4:
            var sa4 syscall.SockaddrInet4
            copy(sa4.Addr[:], taddr.IP[:])
            sa4.Port = taddr.Port
            sa = &sa4
        case 16:
            var sa6 syscall.SockaddrInet6
            copy(sa6.Addr[:], taddr.IP[:])
            sa6.Port = taddr.Port
            sa = &sa6
        }
    }
    return sa, nil
}

// servenet uses the stdlib net package instead of syscalls.
func servenet(events Events, lns []*listener) error {
    var idc int64
    var mu sync.Mutex
    var cmu sync.Mutex
    var idconn = make(map[int]*netConn)
    var done int64
    var shutdown func(err error)

    // connloop handles an individual connection
    connloop := func(id int, conn net.Conn, lnidx int, ln net.Listener) {
        var closed bool
        defer func() {
            if !closed {
                conn.Close()
            }
        }()
        var packet [0xFFFF]byte
        var cout []byte
        var caction Action
        c := &netConn{id: id, conn: conn}
        cmu.Lock()
        idconn[id] = c
        cmu.Unlock()
        if events.Opened != nil {
            var out []byte
            var opts Options
            var action Action
            mu.Lock()
            if atomic.LoadInt64(&done) == 0 {
                out, opts, action = events.Opened(id, Info{
                    AddrIndex:  lnidx,
                    LocalAddr:  conn.LocalAddr(),
                    RemoteAddr: conn.RemoteAddr(),
                })
            }
            mu.Unlock()
            if opts.TCPKeepAlive > 0 {
                if conn, ok := conn.(*net.TCPConn); ok {
                    conn.SetKeepAlive(true)
                    conn.SetKeepAlivePeriod(opts.TCPKeepAlive)
                }
            }
            if len(out) > 0 {
                cout = append(cout, out...)
            }
            caction = action
        }
        for {
            var n int
            var err error
            var out []byte
            var action Action
            if caction != None {
                goto write
            }
            if len(cout) > 0 || atomic.LoadInt64(&c.wake) != 0 {
                conn.SetReadDeadline(time.Now().Add(time.Microsecond))
            } else {
                conn.SetReadDeadline(time.Now().Add(time.Second))
            }
            n, err = c.Read(packet[:])
            if err != nil && !istimeout(err) {
                if err != io.EOF {
                    c.err = err
                }
                goto close
            }

            if n > 0 {
                if events.Data != nil {
                    mu.Lock()
                    if atomic.LoadInt64(&done) == 0 {
                        out, action = events.Data(id, append([]byte{}, packet[:n]...))
                    }
                    mu.Unlock()
                }
            } else if atomic.LoadInt64(&c.wake) != 0 {
                atomic.StoreInt64(&c.wake, 0)
                if events.Data != nil {
                    mu.Lock()
                    if atomic.LoadInt64(&done) == 0 {
                        out, action = events.Data(id, nil)
                    }
                    mu.Unlock()
                }
            }
            if len(out) > 0 {
                cout = append(cout, out...)
            }
            caction = action
            goto write
        write:
            if len(cout) > 0 {
                if events.Prewrite != nil {
                    mu.Lock()
                    if atomic.LoadInt64(&done) == 0 {
                        action = events.Prewrite(id, len(cout))
                    }
                    mu.Unlock()
                    if action == Shutdown {
                        caction = Shutdown
                    }
                }
                conn.SetWriteDeadline(time.Now().Add(time.Microsecond))
                n, err := c.Write(cout)
                if err != nil && !istimeout(err) {
                    if err != io.EOF {
                        c.err = err
                    }
                    goto close
                }
                cout = cout[n:]
                if len(cout) == 0 {
                    cout = nil
                }
                if events.Postwrite != nil {
                    mu.Lock()
                    if atomic.LoadInt64(&done) == 0 {
                        action = events.Postwrite(id, n, len(cout))
                    }
                    mu.Unlock()
                    if action == Shutdown {
                        caction = Shutdown
                    }
                }
            }
            if caction == Shutdown {
                goto close
            }
            if len(cout) == 0 {
                if caction != None {
                    goto close
                }
            }
            continue
        close:
            cmu.Lock()
            delete(idconn, c.id)
            cmu.Unlock()
            mu.Lock()
            if atomic.LoadInt64(&done) != 0 {
                mu.Unlock()
                return
            }
            mu.Unlock()
            if caction == Detach {
                if events.Detached != nil {
                    if len(cout) > 0 {
                        c.outbuf = cout
                    }
                    c.detached = true
                    conn.SetDeadline(time.Time{})
                    mu.Lock()
                    if atomic.LoadInt64(&done) == 0 {
                        caction = events.Detached(c.id, c)
                    }
                    mu.Unlock()
                    closed = true
                    if caction == Shutdown {
                        goto fail
                    }
                    return
                }
            }
            conn.Close()
            if events.Closed != nil {
                var action Action
                mu.Lock()
                if atomic.LoadInt64(&done) == 0 {
                    action = events.Closed(c.id, c.err)
                }
                mu.Unlock()
                if action == Shutdown {
                    caction = Shutdown
                }
            }
            closed = true
            if caction == Shutdown {
                goto fail
            }
            return
        fail:
            shutdown(nil)
            return
        }
    }

    ctx := Server{
        Wake: func(id int) bool {
            cmu.Lock()
            c := idconn[id]
            cmu.Unlock()
            if c == nil {
                return false
            }
            atomic.StoreInt64(&c.wake, 1)
            // force a quick wakeup
            c.conn.SetDeadline(time.Time{}.Add(1))
            return true
        },
        Dial: func(addr string, timeout time.Duration) int {
            if atomic.LoadInt64(&done) != 0 {
                return 0
            }
            id := int(atomic.AddInt64(&idc, 1))
            go func() {
                network, address, _ := parseAddr(addr)
                var conn net.Conn
                var err error
                if timeout > 0 {
                    conn, err = net.DialTimeout(network, address, timeout)
                } else {
                    conn, err = net.Dial(network, address)
                }
                if err != nil {
                    if events.Opened != nil {
                        mu.Lock()
                        _, _, action := events.Opened(id, Info{Closing: true, AddrIndex: -1})
                        mu.Unlock()
                        if action == Shutdown {
                            shutdown(nil)
                            return
                        }
                    }
                    if events.Closed != nil {
                        mu.Lock()
                        action := events.Closed(id, err)
                        mu.Unlock()
                        if action == Shutdown {
                            shutdown(nil)
                            return
                        }
                    }
                    return
                }
                go connloop(id, conn, -1, nil)
            }()
            return id
        },
    }
    var swg sync.WaitGroup
    swg.Add(1)
    var ferr error
    shutdown = func(err error) {
        mu.Lock()
        if atomic.LoadInt64(&done) != 0 {
            mu.Unlock()
            return
        }
        defer swg.Done()
        atomic.StoreInt64(&done, 1)
        ferr = err
        for _, ln := range lns {
            ln.ln.Close()
        }
        type connid struct {
            conn net.Conn
            id   int
        }
        var connids []connid
        cmu.Lock()
        for id, conn := range idconn {
            connids = append(connids, connid{conn.conn, id})
        }
        idconn = make(map[int]*netConn)
        cmu.Unlock()
        mu.Unlock()
        sort.Slice(connids, func(i, j int) bool {
            return connids[j].id < connids[i].id
        })
        for _, connid := range connids {
            connid.conn.Close()
            if events.Closed != nil {
                mu.Lock()
                events.Closed(connid.id, nil)
                mu.Unlock()
            }
        }
    }
    ctx.Addrs = make([]net.Addr, len(lns))
    for i, ln := range lns {
        ctx.Addrs[i] = ln.naddr
    }
    if events.Serving != nil {
        if events.Serving(ctx) == Shutdown {
            return nil
        }
    }
    var lwg sync.WaitGroup
    lwg.Add(len(lns))
    for i, ln := range lns {
        go func(lnidx int, ln net.Listener) {
            defer lwg.Done()
            for {
                conn, err := ln.Accept()
                if err != nil {
                    if err == io.EOF {
                        shutdown(nil)
                    } else {
                        shutdown(err)
                    }
                    return
                }
                id := int(atomic.AddInt64(&idc, 1))
                go connloop(id, conn, lnidx, ln)
            }
        }(i, ln.ln)
    }
    go func() {
        for {
            mu.Lock()
            if atomic.LoadInt64(&done) != 0 {
                mu.Unlock()
                break
            }
            mu.Unlock()
            var delay time.Duration
            var action Action
            mu.Lock()
            if events.Tick != nil {
                if atomic.LoadInt64(&done) == 0 {
                    delay, action = events.Tick()
                }
            } else {
                mu.Unlock()
                break
            }
            mu.Unlock()
            if action == Shutdown {
                shutdown(nil)
                return
            }
            time.Sleep(delay)
        }
    }()
    lwg.Wait() // wait for listeners
    swg.Wait() // wait for shutdown
    return ferr
}

func istimeout(err error) bool {
    if err, ok := err.(net.Error); ok && err.Timeout() {
        return true
    }
    return false
}
