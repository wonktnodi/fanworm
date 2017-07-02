package services

import (
    "sync"
    "../proto"
)

type BucketOptions struct {
    ChannelSize   int
    RoomSize      int
    RoutineAmount int64
    RoutineSize   int
}

// Bucket is a channel holder.
type Bucket struct {
    cLock    sync.RWMutex       // protect the channels for chs
    chs      map[int64]*Session // map sub key to a channel
    boptions BucketOptions
    // room
    rooms map[int64]*Room // bucket room channels
    //routines    []chan *proto.BoardcastRoomArg
    routinesNum uint64
}

func NewBucket(boptions BucketOptions) (b *Bucket) {
    b = new(Bucket)
    b.chs = make(map[int64]*Session, boptions.ChannelSize)
    b.boptions = boptions

    b.rooms = make(map[int64]*Room, boptions.RoomSize)
    //b.routines = make([]chan *proto.BoardcastRoomArg, boptions.RoutineAmount)
    //for i := int64(0); i < boptions.RoutineAmount; i++ {
    //    c := make(chan *proto.BoardcastRoomArg, boptions.RoutineSize)
    //    b.routines[i] = c
    //    go b.roomproc(c)
    //}
    return
}

// Put put a session according with sub key.
func (b *Bucket) PutSession(key string, ch *Session) (err error) {
    return
}

// Del delete the session by sub key.
func (b *Bucket) DelSession(key string) {

}

// Channel get a session by sub key.
func (b *Bucket) Session(key string) (ch *Session) {
    return
}

// Broadcast push msgs to all channels in the bucket.
func (b *Bucket) Broadcast(p *proto.Proto) {

}

func (b *Bucket) Room(rid int32) (room *Room) {
    return
}

func (b *Bucket) DelRoom(rid int32) {

}

// BroadcastRoom broadcast a message to specified room
func (b *Bucket) BroadcastRoom(arg *proto.BoardcastRoomArg) {

}

// Rooms get all room id where online number > 0.
func (b *Bucket) Rooms() (res map[int32]struct{}) {
    return
}

// roomproc
func (b *Bucket) roomproc(c chan *proto.BoardcastRoomArg) {

}