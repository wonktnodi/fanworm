package base

import "github.com/wonktnodi/go-revolver/pkg/log"

type EventHandler interface {
    HandleInput(fd int) error
    HandleOutput(fd int) error
    HandleException(fd int) error
    HandleTimeout(id uint64) error
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