package base

import (
    "time"
    "runtime"
)

var (
    Conf     *Config
    confFile string
)

type Config struct {
    // base section
    PidFile   string   `goconf:"base:pidfile"`
    Dir       string   `goconf:"base:dir"`
    Log       string   `goconf:"base:log"`
    MaxProc   int      `goconf:"base:maxproc"`
    PprofBind []string `goconf:"base:pprof.bind:,"`
    StatBind  []string `goconf:"base:stat.bind:,"`
    ServerId  int32    `goconf:"base:server.id"`
    Debug     bool     `goconf:"base:debug"`
    Whitelist []string `goconf:"base:white.list:,"`
    WhiteLog  string   `goconf:"base:white.log"`
    // tcp
    TCPBind         []string `goconf:"tcp:bind:,"`
    TCPSndbuf       int      `goconf:"tcp:sndbuf:memory"`
    TCPRcvbuf       int      `goconf:"tcp:rcvbuf:memory"`
    TCPKeepalive    bool     `goconf:"tcp:keepalive"`
    TCPReader       int      `goconf:"tcp:reader"`
    TCPReadBuf      int      `goconf:"tcp:readbuf"`
    TCPReadBufSize  int      `goconf:"tcp:readbuf.size"`
    TCPWriter       int      `goconf:"tcp:writer"`
    TCPWriteBuf     int      `goconf:"tcp:writebuf"`
    TCPWriteBufSize int      `goconf:"tcp:writebuf.size"`
    // websocket
    WebsocketBind        []string `goconf:"websocket:bind:,"`
    WebsocketTLSOpen     bool     `goconf:"websocket:tls.open"`
    WebsocketTLSBind     []string `goconf:"websocket:tls.bind:,"`
    WebsocketCertFile    string   `goconf:"websocket:cert.file"`
    WebsocketPrivateFile string   `goconf:"websocket:private.file"`
    // flash safe policy
    FlashPolicyOpen bool     `goconf:"flash:policy.open"`
    FlashPolicyBind []string `goconf:"flash:policy.bind:,"`
    // proto section
    HandshakeTimeout time.Duration `goconf:"proto:handshake.timeout:time"`
    WriteTimeout     time.Duration `goconf:"proto:write.timeout:time"`
    SvrProto         int           `goconf:"proto:svr.proto"`
    CliProto         int           `goconf:"proto:cli.proto"`
    // timer
    Timer     int `goconf:"timer:num"`
    TimerSize int `goconf:"timer:size"`
    // bucket
    Bucket        int   `goconf:"bucket:num"`
    BucketChannel int   `goconf:"bucket:channel"`
    BucketRoom    int   `goconf:"bucket:room"`
    RoutineAmount int64 `goconf:"bucket:routine.amount"`
    RoutineSize   int   `goconf:"bucket:routine.size"`
    // push
    RPCPushAddrs []string `goconf:"push:rpc.addrs:,"`
    // logic
    LogicAddrs []string `goconf:"logic:rpc.addrs:,"`
    // monitor
    MonitorOpen  bool     `goconf:"monitor:open"`
    MonitorAddrs []string `goconf:"monitor:addrs:,"`
}

func NewConfig() *Config {
    return &Config{
        // base section
        PidFile:   "/tmp/goim-comet.pid",
        Dir:       "./",
        Log:       "./comet-log.xml",
        MaxProc:   runtime.NumCPU(),
        PprofBind: []string{"localhost:6971"},
        StatBind:  []string{"localhost:6972"},
        Debug:     true,
        // tcp
        TCPBind:      []string{"0.0.0.0:8080"},
        TCPSndbuf:    1024,
        TCPRcvbuf:    1024,
        TCPKeepalive: false,
        TCPReader:    16,
        TCPWriter:    16,
        // websocket
        WebsocketBind: []string{"0.0.0.0:8090"},
        // websocket tls
        WebsocketTLSOpen:     false,
        WebsocketTLSBind:     []string{"0.0.0.0:8095"},
        WebsocketCertFile:    "../source/cert.pem",
        WebsocketPrivateFile: "../source/private.pem",
        // flash safe policy
        FlashPolicyOpen: false,
        FlashPolicyBind: []string{"0.0.0.0:843"},
        // proto section
        HandshakeTimeout: 5 * time.Second,
        WriteTimeout:     5 * time.Second,
        TCPReadBuf:       1024,
        TCPWriteBuf:      1024,
        TCPReadBufSize:   1024,
        TCPWriteBufSize:  1024,
        // timer
        Timer:     runtime.NumCPU(),
        TimerSize: 1000,
        // bucket
        Bucket:        1024,
        CliProto:      5,
        SvrProto:      80,
        BucketChannel: 1024,
        // push
        RPCPushAddrs: []string{"localhost:8083"},
    }
}

// InitConfig init the global config.
func InitConfig() (err error) {
    Conf = NewConfig()
    return nil
}