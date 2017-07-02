package main

import (
    "fmt"
    "./socket.io"
    "./socket.io/parser"
    "net/http"
    log "./logmate"
    "runtime"
    "./perf"
    . "./services"
)

var logger log.Logger

var (
    //DefaultServer    *Server
    //DefaultWhitelist *Whitelist
    Debug            bool
)

var DefaultServer *Server


func main() {
    logger = log.Start(log.LogFilePath("./log"), log.EveryHour, log.AlsoStdout,
        log.LogFlags(log.Lfunc|log.Lfile|log.Lline))
    defer logger.Stop()

    //startSocketIO()
    startServer()

}

func startServer() {
    if err := InitConfig(); err != nil {
        panic(err)
    }

    Debug = Conf.Debug
    runtime.GOMAXPROCS(Conf.MaxProc)

    perf.Init(Conf.PprofBind)

    // initialize server
    buckets := make([]*Bucket, Conf.Bucket)
    for i := 0; i < Conf.Bucket; i++ {
        buckets[i] = NewBucket(BucketOptions{
            ChannelSize:   Conf.BucketChannel,
            RoomSize:      Conf.BucketRoom,
            RoutineAmount: Conf.RoutineAmount,
            RoutineSize:   Conf.RoutineSize,
        })
    }

    round := NewRound(RoundOptions{
        Reader:       Conf.TCPReader,
        ReadBuf:      Conf.TCPReadBuf,
        ReadBufSize:  Conf.TCPReadBufSize,
        Writer:       Conf.TCPWriter,
        WriteBuf:     Conf.TCPWriteBuf,
        WriteBufSize: Conf.TCPWriteBufSize,
        Timer:        Conf.Timer,
        TimerSize:    Conf.TimerSize,
    })

    operator := new(DefaultOperator)
    DefaultServer = NewServer(buckets, round, operator, ServerOptions{
        CliProto:         Conf.CliProto,
        SvrProto:         Conf.SvrProto,
        HandshakeTimeout: Conf.HandshakeTimeout,
        TCPKeepalive:     Conf.TCPKeepalive,
        TCPRcvbuf:        Conf.TCPRcvbuf,
        TCPSndbuf:        Conf.TCPSndbuf,
    })
}

func startSocketIO() {
    server, err := socketio.NewServer(nil)
    if err != nil {
        log.Fatalln(err)
    }
    server.OnConnect("/", func(s socketio.Conn) error {
        s.SetContext("")
        fmt.Println("connected:", s.ID())
        return nil
    })

    server.OnConnect("/chat", func(s socketio.Conn) error {
        s.SetContext("")
        fmt.Println("chat connected:", s.ID())
        return nil
    })

    server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
        fmt.Println("notice:", msg)
        s.Emit("reply", "have "+msg)
    })

    server.OnEvent("/", "test", func(s socketio.Conn, msg *parser.Buffer) {
        fmt.Printf("test: %+v\n", msg)
    })
    server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
        s.SetContext(msg)
        return "recv " + msg
    })
    server.OnEvent("/", "bye", func(s socketio.Conn) string {
        last := s.Context().(string)
        s.Emit("bye", last)
        s.Close()
        return last
    })
    server.OnError("/", func(e error) {
        fmt.Println("meet error:", e)
    })
    server.OnDisconnect("/", func(s socketio.Conn, msg string) {
        fmt.Println("closed", msg)
    })
    go server.Serve()
    defer server.Close()

    http.Handle("/socket.io/", server)
    http.Handle("/", http.FileServer(http.Dir("./asset")))
    log.Infoln("Serving at localhost:8000...")
    log.Fatalln(http.ListenAndServe(":8000", nil))
}