package main

import (
    "flag"
    "fmt"
    "github.com/wonktnodi/go-revolver/pkg/demo"
    "log"
)

func main()  {
    var port int
    flag.IntVar(&port, "port", 5000, "server port")
    flag.Parse()

    var events demo.Events
    events.Serving = func(srv demo.Server) (action demo.Action) {
        log.Printf("echo server started on port %d", port)
        return
    }
    events.Opened = func(id int, info demo.Info) (out []byte, opts demo.Options, action demo.Action) {
        log.Printf("opened: %d: %s", id, info.RemoteAddr.String())
        return
    }
    events.Closed = func(id int, err error) (action demo.Action) {
        log.Printf("closed: %d", id)
        return
    }
    events.Data = func(id int, in []byte) (out []byte, action demo.Action) {
        out = in
        return
    }
    // parse listen address
    var r = demo.NewReactor()
    r.Open()
    r.Serve(events, fmt.Sprintf("tcp://:%d", port))
}
