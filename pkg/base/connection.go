package base

import (
    "time"
    "net"
    "syscall"
)

// unixConn represents the connection as the event loop sees it.
// This is also becomes a detached connection.
type Connection struct {
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

func (c *Connection) Timeout() time.Time {
    return c.timeout
}

func (c *Connection) Connect(addr string) (err error) {
    sa, err := resolve(addr)
    if nil != err {
        return
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

    c.fd = fd
    c.raddr = sockaddrToAddr(sa)
    filladdrs(c)

    ReactorInstance().AddEventHandler(c, kWriteEvent)
    return
}

func (c *Connection) Read(p []byte) (n int, err error) {
    return syscall.Read(c.fd, p)
}

func (c *Connection) Write(p []byte) (n int, err error) {
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

func (c *Connection) Close() error {
    if c.closed {
        return syscall.EINVAL
    }
    err := syscall.Close(c.fd)
    c.fd = -1
    c.closed = true
    return err
}

func (c *Connection) GetFd() int {
    return c.fd
}

func (c *Connection) HandleInput(fd int) (err error) {
    var packet [1024]byte
    n, err := syscall.Read(fd, packet[:])

    if n == 0 || err != nil {
        if err == syscall.EAGAIN {
            return
        }
        c.err = err
        //TODO: try to close

        return
    }

    //n, err = syscall.Write(fd, packet[:n])

    return
}

func (c *Connection) HandleOutput(fd int) (err error) {
    ReactorInstance().RemoveEventHandler(c, kWriteEvent)
    ReactorInstance().AddEventHandler(c, kReadEvent)
    return
}

func (c *Connection) HandleException(fd int) (err error) {
    return
}

func (c *Connection) HandleTimeout(id uint64) (err error) {
    return
}

func (c *Connection) GetHandle() int {
    return c.fd
}

func (c *Connection) GetID() int {
    return c.id
}

func genAddrs(c *Connection) {
    if c.laddr == nil {
        sa, _ := syscall.Getsockname(c.fd)
        c.laddr = sockaddrToAddr(sa)
    }
    if c.raddr == nil {
        sa, _ := syscall.Getsockname(c.fd)
        c.raddr = sockaddrToAddr(sa)
    }
}

type ConnectionMgr struct {
    fdconn map[int]*Connection
    idconn map[int]*Connection
    id     int
}

var connMgr = NewConnectionMgr()

func NewConnectionMgr() (m *ConnectionMgr) {
    m = &ConnectionMgr{
        fdconn: map[int]*Connection{},
        idconn: map[int]*Connection{},
    }
    return
}

func (m *ConnectionMgr) GetID() int {
    m.id ++
    return m.id
}

func (m *ConnectionMgr) AddConnection(c *Connection) {
    m.idconn[c.id] = c
    if c.fd != 0 {
        m.fdconn[c.fd] = c
    }
}

func (m *ConnectionMgr) RemoveConnection(c *Connection) {
    delete(m.fdconn, c.fd)
    delete(m.idconn, c.id)
}

func (m *ConnectionMgr) GetConnection(fd int) (c *Connection) {
    c = m.fdconn[fd]
    return c
}

func (m *ConnectionMgr) GetConnectionByID(id int) (c *Connection) {
    c = m.idconn[id]
    return c
}
