package demo

import (
    "time"
    "net"
    "syscall"
)

// unixConn represents the connection as the event loop sees it.
// This is also becomes a detached connection.
type connection struct {
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

func (c *connection) Timeout() time.Time {
    return c.timeout
}
func (c *connection) Read(p []byte) (n int, err error) {
    return syscall.Read(c.fd, p)
}
func (c *connection) Write(p []byte) (n int, err error) {
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

func (c *connection) Close() error {
    if c.closed {
        return syscall.EINVAL
    }
    err := syscall.Close(c.fd)
    c.fd = -1
    c.closed = true
    return err
}

func genAddrs(c *connection) {
    if c.laddr == nil {
        sa, _ := syscall.Getsockname(c.fd)
        c.laddr = sockaddrToAddr(sa)
    }
    if c.raddr == nil {
        sa, _ := syscall.Getsockname(c.fd)
        c.raddr = sockaddrToAddr(sa)
    }
}
