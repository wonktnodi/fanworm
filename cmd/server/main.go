package main

import (
    "github.com/wonktnodi/go-revolver/pkg/log"
    "github.com/wonktnodi/go-revolver/pkg/base"
    "flag"
    "fmt"
)

func main() {
    defer log.Start(log.LogFilePath("./log"), log.AlsoStdout, log.EveryDay).Stop()
    log.Infoln("start service...")

    var port int
    flag.IntVar(&port, "port", 5000, "server port")
    flag.Parse()

    var events base.Events
    events.Serving = func(srv base.Server) (action base.Action) {
        log.Debugf("echo server started on port %d", port)
        return
    }
    events.Opened = func(id int, info base.Info) (out []byte, opts base.Options, action base.Action) {
        log.Debugf("opened: %d: %s", id, info.RemoteAddr.String())
        return
    }
    events.Closed = func(id int, err error) (action base.Action) {
        log.Debugf("closed: %d", id)
        return
    }
    events.Data = func(id int, in []byte) (out []byte, action base.Action) {
        out = in
        return
    }
    log.Fatalln(base.Serve(events, fmt.Sprintf("tcp://:%d", port)))
}
