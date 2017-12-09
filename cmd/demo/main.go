package main

import (
    "fmt"
    "github.com/wonktnodi/go-revolver/pkg/log"
    "net"
    "syscall"
    "time"
    "github.com/wonktnodi/go-revolver/pkg/base"
)

func main() {
    l := log.Start(log.EveryDay)
    defer l.Stop()

    base.InitPollReactor()

    kqueue_run()
}

func kqueue_run() {
    var err error

    err = base.ReactorInstance().Open()
    //epollfd, err := syscall.Kqueue()
    if err != nil {
        return
    }

    addr := "tcp://0.0.0.0:5000"
    var ln = base.NewTcpListener()
    err = ln.Open(addr)
    if err != nil {
        fmt.Println("exit by error, ", err)
        return
    }

    for {
        time.Sleep(time.Hour)
    }
    //updateEvents(epollfd, ln.GetHandle(), kReadEvent, false)

    //for {
    //    loop_once(epollfd, ln.GetHandle())
    //}

}

func loop_once(efd, lfd int) {
    ts := syscall.NsecToTimespec(int64(10 * time.Second))
    evs := make([]syscall.Kevent_t, 20)

    n, _ := syscall.Kevent(efd, nil, evs, &ts)
    for i := 0; i < n; i ++ {
        fd := int(evs[i].Ident)
        events := evs[i].Filter
        if events == syscall.EVFILT_READ {
            if fd == lfd {
                handleAccept(efd, lfd)
            } else {
                handleRead(efd, fd)
            }
        } else if events == syscall.EVFILT_WRITE {
            handleWrite(efd, fd)
        }
    }
}

// unixConn represents the connection as the event loop sees it.
// This is also becomes a detached connection.
type unixConn struct {
    id, fd int
    outbuf []byte
    outpos int
    //action   Action
    //opts     Options
    timeout  time.Time
    raddr    net.Addr // remote addr
    laddr    net.Addr // local addr
    lnidx    int
    err      error
    dialerr  error
    wake     bool
    readon   bool
    writeon  bool
    detached bool
    closed   bool
    opening  bool
}

func handleAccept(efd, fd int) {
    cfd, rsa, err := syscall.Accept(fd)
    if err != nil {
        log.Traceln(rsa)
        return
    }

    if err = syscall.SetNonblock(cfd, true); err != nil {
        return
    }

    //c := &unixConn{id: 0, fd: cfd,
    //    opening: true,
    //    lnidx:   0,
    //    raddr:   sockaddrToAddr(rsa),
    //}

    //updateEvents(efd, cfd, kWriteEvent, false)

}

func handleRead(efd, fd int) {
    var packet [1024]byte
    n, err := syscall.Read(fd, packet[:])

    if n == 0 || err != nil {
        if err == syscall.EAGAIN {
            return
        }
        //c.err = err
        //goto close
    }
    n, err = syscall.Write(fd, packet[:n])
}

func handleWrite(efd, fd int) {
    //updateEvents(efd, fd, kReadEvent, true);
}


//func reactor_run() {
//    var port int
//    flag.IntVar(&port, "port", 5000, "server port")
//    flag.Parse()
//
//    var events demo.Events
//    events.Serving = func(srv demo.Server) (action demo.Action) {
//        log.Debugf("echo server started on port %d", port)
//        return
//    }
//    events.Opened = func(id int, info demo.Info) (out []byte, opts demo.Options, action demo.Action) {
//        log.Debugf("opened: %d: %s", id, info.RemoteAddr.String())
//        return
//    }
//    events.Closed = func(id int, err error) (action demo.Action) {
//        log.Debugf("closed: %d", id)
//        return
//    }
//    events.Data = func(id int, in []byte) (out []byte, action demo.Action) {
//        out = in
//        return
//    }
//    // parse listen address
//    var r = demo.NewReactor()
//    r.Open()
//    r.Serve(events, fmt.Sprintf("tcp://:%d", port))
//}
