package base

import "github.com/wonktnodi/go-revolver/pkg/log"

//type EventHandler struct {
//    // Opened fires when a new connection has opened.
//    // The info parameter has information about the connection such as
//    // it's local and remote address.
//    // Use the out return value to write data to the connection.
//    // The opts return value is used to set connection options.
//    OnConnected func(id int, info Info) (out []byte, opts Options, action Action)
//
//    // Closed fires when a connection has closed.
//    // The err parameter is the last known connection error.
//    Closed func(id int, err error) (action Action)
//
//    // Tick fires immediately after the server starts and will fire again
//    // following the duration specified by the delay return value.
//    OnTimeout func() (delay time.Duration, action Action)
//}

type EventHandler interface {
    HandleInput(fd int) error
    HandleOutput(fd int) error
    HandleException(fd int) error
    HandleTimeout(fd int) error
    GetHandle() int
    GetID() int
}

type EventBase struct {
    ID int
    fd int
}

type HandlerMgr struct {
    sequence int
    handlers  map[int]EventHandler
}

func NewHandlerMgr() (mgr *HandlerMgr, err error) {
    mgr = &HandlerMgr{}
    mgr.Init()
    return
}

func (h *HandlerMgr) Init() {
    h.sequence = 100
    h.handlers = make(map[int]EventHandler)
}

func (h *HandlerMgr) GetHandler(fd int) (handler EventHandler) {
    var ok bool
    handler, ok = h.handlers[fd]
    if ok == false {
        log.Warnf("failed to get event handler for fd[%d]", fd)
    }
    return
}

func (h *HandlerMgr) AddEvent(ev EventHandler) (err error) {
    id := (ev).GetID()
    if _, ok := h.handlers[id]; ok {
        log.Errorf("add duplicated event for ID[%d]", id)
    }
    h.handlers[id] = ev
    return
}

func (h *HandlerMgr) RemoveEvent(ev EventHandler) (err error){
    id := (ev).GetID()
    if _, ok := h.handlers[id]; !ok {
        log.Errorf("add duplicated event for ID[%d]", id)
        return
    }
    delete(h.handlers, id)
    return
}

func (h *HandlerMgr) GenerateID() int {
    id := h.sequence
    h.sequence ++
    return id
}