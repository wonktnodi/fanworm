package base

type ReactorEventHandlerInfo struct {
    eventHandler EventHandler
    eventMask    uint32
    eventClose   bool
}

var eventHandlers = map[int]ReactorEventHandlerInfo{}


func (r *PollReactor) registerEventHandler(handler EventHandler, masks uint32) (err error) {

    return
}

func (r *PollReactor) removeEventHandler(handler EventHandler, mask uint32) {
    ev, ok := eventHandlers[handler.GetHandle()]
    if ok == false {
        return
    }

    ev.eventMask = ev.eventMask &^ mask
}

func (r *PollReactor) deleteEventHandler(handler EventHandler) {
    delete(eventHandlers, handler.GetHandle())
    // TODO: clean the file descriptor event watching ?
}