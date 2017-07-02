package services

import (
    "bufio"
)

type Session struct {
    RoomID int64
    Writer bufio.Writer
    Reader bufio.Reader
    Next   *Session
    Prev   *Session
}

func NewSession(cli, svr int, rid int64) *Session {
    c := new(Session)
    c.RoomID = rid
    //c.CliProto.Init(cli)
    //c.signal = make(chan *proto.Proto, svr)
    return c
}

// Push server push message.
func (c *Session) Push(/*p *proto.Proto*/) (err error) {
    return
}

// Ready check the channel ready or close?
func (c *Session) Ready()  {

}

// Signal send signal to the channel, protocol ready.
func (c *Session) Signal() {
    //c.signal <- proto.ProtoReady
}

// Close close the channel.
func (c *Session) Close() {
    //c.signal <- proto.ProtoFinish
}
