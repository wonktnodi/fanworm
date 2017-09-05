package base

import (
    "time"
    log "github.com/wonktnodi/go-revolver/utils/logmate"
    "github.com/wonktnodi/go-revolver/utils/hash/cityhash"
)

var (
    maxInt        = 1<<31 - 1
    emptyJSONBody = []byte("{}")
)

type ServerOptions struct {
    CliProto         int
    SvrProto         int
    HandshakeTimeout time.Duration
    TCPKeepalive     bool
    TCPRcvbuf        int
    TCPSndbuf        int
}

type Server struct {
    Options   ServerOptions
    Buckets   []*Bucket // subkey bucket
    bucketIdx uint32
    round     *Round // accept round store
    operator  Operator
}

// NewServer returns a new Server.
func NewServer(options ServerOptions, o Operator) *Server {
    s := new(Server)
    s.Options = options

    // new server
    buckets := make([]*Bucket, Conf.Bucket)
    for i := 0; i < Conf.Bucket; i++ {
        buckets[i] = NewBucket(BucketOptions{
            ChannelSize:   Conf.BucketChannel,
            RoomSize:      Conf.BucketRoom,
            RoutineAmount: Conf.RoutineAmount,
            RoutineSize:   Conf.RoutineSize,
        })
    }
    s.Buckets = buckets
    s.bucketIdx = uint32(len(buckets))

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
    s.round = round

    s.operator = o
    return s
}

func (server *Server) Bucket(subKey string) *Bucket {
    idx := cityhash.CityHash32([]byte(subKey), uint32(len(subKey))) % server.bucketIdx

    log.Tracef("\"%s\" hit channel bucket index: %d use cityhash", subKey, idx)

    return server.Buckets[idx]
}
