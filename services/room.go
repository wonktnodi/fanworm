package services

import "sync"

type Room struct {
    id     int64
    rLock  sync.RWMutex
    next   *Session
    drop   bool
    Online int // dirty read is ok
}

// NewRoom new a room struct, store channel room info.
func NewRoom(id int64) (r *Room) {
    r = new(Room)
    r.id = id
    r.drop = false
    r.next = nil
    r.Online = 0
    return
}

func (r *Room) PutSession(ch *Session) (err error) {
    return
}

func (r *Room) DelSession(ch *Session) (err error) {
    return
}

func (r *Room) Push() {

}

func (r *Room) Close() {

}
