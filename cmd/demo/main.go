package main

import (
    "fmt"
    "github.com/wonktnodi/go-revolver/pkg/log"
    "time"
    "github.com/wonktnodi/go-revolver/pkg/base"
)

func main() {
    l := log.Start(log.EveryDay)
    defer l.Stop()

    base.InitPollReactor()

    kqueue_run()
}

var ln = base.NewTcpListener()

func kqueue_run() {
    var err error

    err = base.ReactorInstance().Open()
    //epollfd, err := syscall.Kqueue()
    if err != nil {
        return
    }

    //addr := "tcp://0.0.0.0:5000"
    //Listen(addr)

    conn := base.Connection{}
    conn.Connect("tcp://127.0.0.1:5000")

    for {
        conn.Write([]byte("11111"))
        time.Sleep(time.Second * 20)
    }
}

func Listen(addr string) (err error) {
    err = ln.Open(addr)
    if err != nil {
        fmt.Println("exit by error, ", err)
        return
    }
    return
}

