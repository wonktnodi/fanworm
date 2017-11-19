package base

import (
    "net/http"
    "github.com/wonktnodi/go-revolver/pkg/log"
    "time"
    "sync/atomic"
)

type Reactor interface {
    AddEventHandler(handler *EventHandler) (err error)
    RemoveEventHandler(handler *EventHandler) (err error)
    DeleteEventHandler(handler *EventHandler) (err error)
    Open() (err error)
    Close()
    EventLoop()
    SetTimer(ev EventHandler, delay int64, data interface{}) (id uint64, err error)
    CancelTimer(id uint64) (err error)
}

type timerInfo struct {
    id      uint64
    timeout time.Time
    handler EventHandler
    data    interface{}
}

func (t timerInfo) Timeout() time.Time {
    return t.timeout
}

func (t timerInfo) TimerID() uint64 {
    return t.id
}


type PollReactor struct {
    PollBase
    pollerFD   int
    timerQueue *TimeoutQueue
    timerID    uint64
    eventsMgr  *HandlerMgr

}

func NewPollReactor() (p *PollReactor, err error) {
    p = &PollReactor{}
    err = p.Open()
    if err != nil {
        log.Errorf("failed to create event handler mgr, err: %v", err)
        p = nil
        return
    }
    return
}

func (r *PollReactor) Open() (err error) {
    r.pollerFD, err = CreatePoll()
    if err != nil {
        log.Errorf("failed to create poll, err:%v", err)
        return
    }
    r.timerQueue = NewTimeoutQueue()
    r.eventsMgr, err = NewHandlerMgr()
    if err != nil {
        log.Errorf("failed to create event handler mgr, err: %v", err)
        return
    }
    r.createEvents(128)
    r.timerID = 1

    return
}

func (r *PollReactor) Close() {
    r.timerQueue = nil
    r.eventsMgr = nil
    r.events = nil
    if err := closePoll(r.pollerFD); err != nil {
        log.Errorf("failed to close poll FD, err: %v", err)
    }
}

func (r *PollReactor) EventLoop() {

}

func (r *PollReactor) AddEventHandler(handler *EventHandler) (err error) {
    return
}

func (r *PollReactor) RemoveEventHandler(handler *EventHandler) (err error) {
    return
}

func (r *PollReactor) DeleteEventHandler(handler *EventHandler) (err error) {
    return
}

func (r *PollReactor) SetTimer(ev EventHandler, delay int64, data interface{}) (id uint64, err error) {
    var item timerInfo
    item.handler = ev
    item.timeout = time.Now().Add(time.Duration(delay))
    item.data = data
    item.id = atomic.LoadUint64(&r.timerID)
    atomic.AddUint64(&r.timerID, 1)
    r.timerQueue.Push(item)
    id = item.id
    return
}

func (r *PollReactor) CancelTimer(id uint64) (err error) {
    return r.timerQueue.Remove(id)
}

func (r *PollReactor) RegisterHandler(handler EventHandler) (err error) {
    return
}

func (r *PollReactor) RemoveHandler(handler http.Handler) (err error) {
    return
}

func (r *PollReactor) Serve() (err error) {
    return
}
