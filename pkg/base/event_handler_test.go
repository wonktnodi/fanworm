package base

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "time"
)

type TestEventHandler struct{
    EventBase
    timeout time.Time
}

func (h TestEventHandler) HandleInput(fd int) (err error) {
    return
}

func (h TestEventHandler) HandleOutput(fd int) (err error) {
    return
}

func (h TestEventHandler) HandleException(fd int) (err error) {
    return
}

func (h TestEventHandler) HandleTimeout(fd int) (err error) {
    return
}

func (h TestEventHandler) GetHandle() (id int) {
    return h.fd
}

func (h TestEventHandler) GetID() (id int) {
    return h.ID
}

func (h TestEventHandler) Timeout() (tm time.Time) {
    return h.timeout
}



func TestHandlerMgr(t *testing.T) {
    mgr, err := NewHandlerMgr()
    assert.Equal(t, nil, err, "create handler from mgr")

    var handler TestEventHandler
    handler.ID = mgr.GenerateID()
    assert.Equal(t, 100, handler.ID)
    assert.Equal(t, 100, handler.GetID())
    assert.Equal(t, 101, mgr.sequence)

    err = mgr.AddEvent(&handler)
    assert.Equal(t, nil, err, "add handler from mgr")
    assert.Equal(t, 1, len(mgr.handlers))

    mgr.RemoveEvent(&handler)
    assert.Equal(t, nil, err, "remove handler from mgr")
    assert.Equal(t, 0, len(mgr.handlers))
}

