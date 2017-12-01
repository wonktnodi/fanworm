package base

import (
    "net"
    "os"
    "syscall"
)

type TcpAcceptor struct {
    ln       net.Listener
    f        *os.File
    fd       int
    network  string
    addr     string
    inetAddr net.Addr
}

func NewTcpAcceptor() (a *TcpAcceptor) {
    a = &TcpAcceptor{}
    return
}

func (ln *TcpAcceptor) GetFD() int {
    return ln.fd
}

func (ln *TcpAcceptor) Open(addr string) (err error) {
    //var stdlib bool
    ln.network, ln.addr, _ = parseAddr(addr)
    if ln.fd != 0 {
        syscall.Close(ln.fd)
    }

    if ln.network == "unix" {
        os.RemoveAll(ln.addr)
    }

    ln.ln, err = net.Listen(ln.network, ln.addr)
    if err != nil {
        return err
    }
    ln.inetAddr = ln.ln.Addr()
    if ln.f != nil {
        ln.f.Close()
    }
    if ln.ln != nil {
        ln.ln.Close()
    }
    if ln.network == "unix" {
        os.RemoveAll(ln.addr)
    }
    return
}

func (ln *TcpAcceptor) system() (err error) {
    switch netln := ln.ln.(type) {
    default:
        panic("invalid listener type")
    case *net.TCPListener:
        ln.f, err = netln.File()
    case *net.UnixListener:
        ln.f, err = netln.File()
    }
    if err != nil {
        ln.Close()
        return err
    }
    ln.fd = int(ln.f.Fd())
    return syscall.SetNonblock(ln.fd, true)
}

func (ln *TcpAcceptor) Close() {
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
