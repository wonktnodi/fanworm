package base

import (
    "net"
    "os"
    "syscall"
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
