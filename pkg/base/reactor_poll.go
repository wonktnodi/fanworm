package base

import (
    "sync"
    "github.com/wonktnodi/go-revolver/pkg/log"
    "sync/atomic"
    "syscall"
)

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

func InitPollReactor() {
    NewReactor = newPollReactor
}

func newPollReactor() (p Reactor, err error) {
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
    r.pollerFD, err = createPoll()
    if err != nil {
        log.Errorf("failed to create poll, err:%v", err)
        return
    }
    //r.createEvents(128)
    r.timerID = 1
    //r.defaultWaitInterval = pkg.MinReactorLoopInterval * 1000
    r.exitFlag = 0

    r.Start()
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
    //r.events = nil
}

func (r *PollReactor) EventLoop() {
    r.wgEnd.Add(1)
    r.wgStart.Done()

    for {
        //ts := syscall.NsecToTimespec(int64(10 * time.Second))
        evs := make([]syscall.Kevent_t, 20)

        n, _ := syscall.Kevent(r.pollerFD, nil, evs, nil)
        var evInfo ReactorEventHandlerInfo
        var ok bool
        for i := 0; i < n; i ++ {
            fd := int(evs[i].Ident)

            if evInfo, ok = eventHandlers[fd]; !ok {
                continue
            }

            events := evs[i].Filter
            if events == syscall.EVFILT_READ {
                evInfo.eventHandler.HandleInput(fd)
            } else if events == syscall.EVFILT_WRITE {
                evInfo.eventHandler.HandleOutput(fd)
            }

            //events := evs[i].Filter
            //if events == syscall.EVFILT_READ {
            //    //if fd == lfd {
            //    //    handleAccept(efd, lfd)
            //    //} else {
            //    //    handleRead(efd, fd)
            //    //}
            //} else if events == syscall.EVFILT_WRITE {
            //    //handleWrite(efd, fd)
            //}
        }
    }

    r.wgEnd.Done()
}

func (r *PollReactor) AddEventHandler(handler EventHandler, masks uint32) (err error) {
    var evInfo ReactorEventHandlerInfo
    var ok bool
    if -1 == r.pollerFD || handler.GetHandle() == -1 {
        return NewError(ErrCodeNotInvalidParam)
    }

    evInfo, ok = eventHandlers[handler.GetHandle()]
    if ok == false {
        evInfo.eventMask = masks
        evInfo.eventClose = false
        evInfo.eventHandler = handler
        err = updateEvents(r.pollerFD, handler.GetHandle(), evInfo.eventMask, false)
    } else {
        evInfo.eventMask = evInfo.eventMask | masks
        err = updateEvents(r.pollerFD, handler.GetHandle(), evInfo.eventMask, true)
    }
    if err != nil {
        return
    }

    eventHandlers[handler.GetHandle()] = evInfo
    return
}

func (r *PollReactor) RemoveEventHandler(handler EventHandler, masks uint32) (err error) {
    var evInfo ReactorEventHandlerInfo
    var ok bool
    evInfo, ok = eventHandlers[handler.GetHandle()]
    if false == ok {
        return NewError(ErrCodeNotFound)
    }
    evInfo.eventMask = evInfo.eventMask &^ masks
    eventHandlers[handler.GetHandle()] = evInfo
    if err = updateEvents(r.pollerFD, handler.GetHandle(), evInfo.eventMask, true); err != nil {
        return
    }

    return
}

func (r *PollReactor) DeleteEventHandler(handler EventHandler, masks uint32) (err error) {
    if -1 == r.pollerFD {
        return NewError(ErrCodeNotInvalidParam)
    }

    //var evInfo ReactorEventHandlerInfo
    var ok bool
    _, ok = eventHandlers[handler.GetHandle()]
    if false == ok {
        return NewError(ErrCodeNotFound)
    }

    if err = updateEvents(r.pollerFD, handler.GetHandle(), 0, true); err != nil {
        return
    }

    delete(eventHandlers, handler.GetHandle())
    return
}

func (r *PollReactor) SetTimer(ev EventHandler, delay int64, data interface{}) (id uint64, err error) {
    return
}

func (r *PollReactor) CancelTimer(id uint64) (err error) {
    return
}
