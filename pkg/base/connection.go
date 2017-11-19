package base

import (
    "time"
    "net"
    "syscall"
)

// Action is an action that occurs after the completion of an event.
type Action int

const (
    // None indicates that no action should occur following an event.
    None Action = iota
    // Detach detaches the client.
    Detach
    // Close closes the client.
    Close
    // Shutdown shutdowns the server.
    Shutdown
)

// Options are set when the client opens.
type Options struct {
    // TCPKeepAlive (SO_KEEPALIVE) socket option.
    TCPKeepAlive time.Duration
}

type unixConn struct {
    id, fd   int
    outbuf   []byte
    outpos   int
    action   Action
    opts     Options
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

func (c *unixConn) Timeout() time.Time {
    return c.timeout
}

func (c *unixConn) TimerID() uint64 {
    return 0
}

func (c *unixConn) Read(p []byte) (n int, err error) {
    return syscall.Read(c.fd, p)
}

func (c *unixConn) Write(p []byte) (n int, err error) {
    if c.detached {
        if len(c.outbuf) > 0 {
            for len(c.outbuf) > 0 {
                n, err = syscall.Write(c.fd, c.outbuf)
                if n > 0 {
                    c.outbuf = c.outbuf[n:]
                }
                if err != nil {
                    return 0, err
                }
            }
            c.outbuf = nil
        }
        var tn int
        if len(p) > 0 {
            for len(p) > 0 {
                n, err = syscall.Write(c.fd, p)
                if n > 0 {
                    p = p[n:]
                    tn += n
                }
                if err != nil {
                    return tn, err
                }
            }
            p = nil
        }
        return tn, nil
    }
    return syscall.Write(c.fd, p)
}

func (c *unixConn) Close() error {
    if c.closed {
        return syscall.EINVAL
    }
    err := syscall.Close(c.fd)
    c.fd = -1
    c.closed = true
    return err
}

func sockaddrToAddr(sa syscall.Sockaddr) net.Addr {
    var a net.Addr
    switch sa := sa.(type) {
    case *syscall.SockaddrInet4:
        a = &net.TCPAddr{
            IP:   append([]byte{}, sa.Addr[:]...),
            Port: sa.Port,
        }
    case *syscall.SockaddrInet6:
        var zone string
        if sa.ZoneId != 0 {
            if ifi, err := net.InterfaceByIndex(int(sa.ZoneId)); err == nil {
                zone = ifi.Name
            }
        }
        if zone == "" && sa.ZoneId != 0 {
        }
        a = &net.TCPAddr{
            IP:   append([]byte{}, sa.Addr[:]...),
            Port: sa.Port,
            Zone: zone,
        }
    case *syscall.SockaddrUnix:
        a = &net.UnixAddr{Net: "unix", Name: sa.Name}
    }
    return a
}

func filladdrs(c *unixConn) {
    if c.laddr == nil {
        sa, _ := syscall.Getsockname(c.fd)
        c.laddr = sockaddrToAddr(sa)
    }
    if c.raddr == nil {
        sa, _ := syscall.Getsockname(c.fd)
        c.raddr = sockaddrToAddr(sa)
    }
}

type Connection struct {
    id, fd  int
    outbuf  []byte
    outpos  int
    action  Action
    opts    Options
    raddr   net.Addr // remote addr
    laddr   net.Addr // local addr
    lnidx   int
    err     error
    dialerr error
}

func (c *Connection) HandleInput(fd int) (err error) {
    return
}

func (c *Connection) HandleOutput(fd int) (err error) {
    return
}

func (c *Connection) HandleException(fd int) (err error) {
    return
}

func (c *Connection) HandleTimeout(fd int) (err error) {
    return
}

func (c *Connection) GetID() (id int) {
    return
}