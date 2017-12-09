package base

import (
    "syscall"
    "github.com/wonktnodi/go-revolver/pkg/log"
)

type TcpListener struct {
    acceptor *TcpAcceptor
}

func NewTcpListener() (l *TcpListener) {
    l = &TcpListener{}

    return
}

func (l *TcpListener) Open(addr string) (err error){
    l.acceptor = NewTcpAcceptor()
    if nil == l.acceptor {
        return NewError(ErrCodeAllocateObject)
    }

    if err = l.acceptor.Open(addr); err != nil {
        return
    }

    return ReactorInstance().AddEventHandler(l, kReadEvent)
}

func (l *TcpListener) Close() {
    if l.acceptor != nil {
        l.acceptor.Close()
    }
}

func (l *TcpListener) GetHandle() int {
    return l.acceptor.fd
}

func (l *TcpListener) HandleInput(fd int) (err error) {
    cfd, rsa, err := syscall.Accept(fd)
    if err != nil {
        log.Traceln(rsa)
        return
    }

    if err = syscall.SetNonblock(cfd, true); err != nil {
        return
    }

    conn := &connection{id: connMgr.GetID(), fd: cfd,
        opening: true,
        lnidx: 0,
        raddr: sockaddrToAddr(rsa),
    }
    connMgr.AddConnection(conn)

    if err = ReactorInstance().AddEventHandler(conn, kWriteEvent); err != nil  {
        return
    }

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

func (l *TcpListener) GetID() int {
    return 0
}