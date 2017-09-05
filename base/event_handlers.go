package base

import (
    "time"
    "github.com/wonktnodi/go-revolver/proto"
    //log "github.com/wonktnodi/go-revolver/utils/logmate"
)

func connect(p *proto.Proto) (key string, rid int32, heartbeat time.Duration, err error) {
    var (
        //arg   = proto.ConnArg{Token: string(p.Body), Server: Conf.ServerId}
        reply = proto.ConnReply{}
    )

    key = reply.Key
    rid = reply.RoomId
    heartbeat = 5 * 60 * time.Second
    return
}

func disconnect(key string, roomId int32) (has bool, err error) {
    var (
        //arg   = proto.DisconnArg{Key: key, RoomId: roomId}
        reply = proto.DisconnReply{}
    )

    has = reply.Has
    return
}
