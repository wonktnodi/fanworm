package demo

import (
    "syscall"
)

type ReactorEventHandlerInfo struct {
    eventHandler EventHandler
    eventMask    uint32
    eventClose   bool
}

var eventHandlers = map[int]ReactorEventHandlerInfo{}

func (r *PollingReactor) registerEventHandler(handler EventHandler, masks uint32) (err error) {
    var evInfo ReactorEventHandlerInfo

    evInfo.eventClose = false
    evInfo.eventHandler = handler
    evInfo.eventMask = masks

    eventHandlers[handler.GetHandle()] = evInfo
    //return base.AddRead(r.p, handler.GetHandle(), nil, nil)

    if masks&EventRead == EventRead {
        _, err = syscall.Kevent(r.p, []syscall.Kevent_t{{
            Ident:  uint64(handler.GetHandle()),
            Flags:  syscall.EV_ADD,
            Filter: syscall.EVFILT_READ}}, nil, nil)
        if err != nil {
            return err
        }
    }
    if masks&EventWrite == EventWrite {
        _, err = syscall.Kevent(r.p, []syscall.Kevent_t{{
            Ident:  uint64(handler.GetHandle()),
            Flags:  syscall.EV_ADD,
            Filter: syscall.EVFILT_WRITE}}, nil, nil)

        if err != nil {
            return err
        }
    }
    return
}

func (r *PollingReactor) removeEventHandler(handler EventHandler, mask uint32) {
    ev, ok := eventHandlers[handler.GetHandle()]
    if ok == false {
        return
    }

    ev.eventMask = ev.eventMask &^ mask
}

func (r *PollingReactor) deleteEventHandler(handler EventHandler) {
    delete(eventHandlers, handler.GetHandle())
    // TODO: clean the file descriptor event watching ?
}
