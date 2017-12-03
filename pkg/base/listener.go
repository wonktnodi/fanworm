package base

import (
    "net"
    "os"
    "syscall"
    "github.com/wonktnodi/go-revolver/pkg/log"
)

type listener struct {
    ln      net.Listener
    f       *os.File
    fd      int
    network string
    addr    string
    naddr   net.Addr
}

func (ln *listener) close() {
    if ln.fd != 0 {
        syscall.Close(ln.fd)
    }
    if ln.f != nil {
        ln.f.Close()
    }
    if ln.ln != nil {
        ln.ln.Close()
    }
    if ln.network == "unix" {
        os.RemoveAll(ln.addr)
    }
}

// system takes the net listener and detaches it from it's parent
// event loop, grabs the file descriptor, and makes it non-blocking.
func (ln *listener) system() error {
    var err error
    switch netln := ln.ln.(type) {
    default:
        panic("invalid listener type")
    case *net.TCPListener:
        ln.f, err = netln.File()
    case *net.UnixListener:
        ln.f, err = netln.File()
    }
    if err != nil {
        ln.close()
        return err
    }
    ln.fd = int(ln.f.Fd())
    return syscall.SetNonblock(ln.fd, true)
}

type TcpListener struct {
    acceptor *TcpAcceptor
}

func NewTcpListener() (l *TcpListener) {
    l = &TcpListener{}
    return
}

func (l *TcpListener) Open(reactor Reactor, addr string) (err error) {
    acceptor := NewTcpAcceptor()
    if acceptor == nil {
        log.Errorf("failed to create acceptor")
        return
    }
    err = acceptor.Open(addr)
    if err != nil {
        return err
    }
    l.acceptor = acceptor

    // register events
    //err = reactor.AddEventHandler(l)
    reactor.AddEventHandler(l, EventRead)
    return
}

func (l *TcpListener) Close() {
    if l.acceptor != nil {
        l.acceptor.Close()
        l.acceptor = nil
    }
}


func (l *TcpListener) HandleInput(fd int) (err error) {
    return
}

func (l *TcpListener) HandleOutput(fd int) (err error) {
    return
}

func (l *TcpListener) HandleException(fd int) (err error) {
    return
}

func (l *TcpListener) HandleTimeout(id uint64) (err error) {
    return
}

func (l *TcpListener) GetHandle() int {
    return l.acceptor.GetFD()
}

func (l *TcpListener) GetID() int {
    return 0
}
