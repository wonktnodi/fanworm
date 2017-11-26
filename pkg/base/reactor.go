package base

import (
    "net/http"
    "github.com/wonktnodi/go-revolver/pkg/log"
    "time"
    "sync/atomic"
    "syscall"
    "github.com/wonktnodi/go-revolver/pkg"
    "sync"
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

func (t timerInfo) HandleTimeout(id uint64) error {
    if nil != t.handler {
        return t.handler.HandleTimeout(t.id)
    } else {
        return nil
    }
}

type PollReactor struct {
    PollBase
    exitFlag            int64
    timerID             uint64
    timerQueue          *TimeoutQueue
    eventsMgr           *HandlerMgr
    defaultWaitInterval int64
    wgStart, wgEnd      sync.WaitGroup
    pollerFD            int
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
    r.defaultWaitInterval = pkg.MinReactorLoopInterval
    r.exitFlag = 0
    return
}

func (r *PollReactor) Start() {
    r.wgStart.Add(1)
    go r.EventLoop()
    r.wgStart.Wait()
}

func (r *PollReactor) Close() {
    atomic.StoreInt64(&r.exitFlag, 1)
    if err := closePoll(r.pollerFD); err != nil {
        log.Errorf("failed to close poll FD, err: %v", err)
    }
    r.wgEnd.Wait()

    r.timerQueue = nil
    r.eventsMgr = nil
    r.events = nil
}

func (r *PollReactor) EventLoop() {
    var now time.Time
    var exitFlag = atomic.LoadInt64(&r.exitFlag)
    r.wgEnd.Add(1)
    r.wgStart.Done()
    for {
        pn, err := r.wait(int64(r.defaultWaitInterval))
        if err != nil && err != syscall.EINTR {
            break
        }
        // event processing
        for i := 0; i < pn; i ++ {

        }

        // check exit flag
        exitFlag = atomic.LoadInt64(&r.exitFlag)
        if exitFlag > 0 {
            break
        }

        // timer processing
        if r.timerQueue.Len() > 0 {
            now = time.Now()
            for {
                v := r.timerQueue.Peek()
                if v == nil {
                    break
                }
                if now.After(v.Timeout()) {
                    r.timerQueue.Pop()
                    v.HandleTimeout(v.TimerID())
                } else {
                    break
                }
            }
        }
    }
    r.wgEnd.Done()
    return
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
