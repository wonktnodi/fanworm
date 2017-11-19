package base

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "time"
)

func TestReactorOpenClose(t *testing.T) {
    // create and initialize testing reactor
    reactor, err := NewPollReactor()
    assert.Equal(t, nil, err)
    assert.NotEqual(t, 0, reactor.pollerFD)
    assert.NotEqual(t, nil, reactor.timerQueue)
    assert.NotEqual(t, nil, reactor.eventsMgr)
    assert.NotEqual(t, nil, reactor.events)

    reactor.EventLoop()

    // try to close reactor
    reactor.Close()
    assert.Nil(t, reactor.events, "events clear")
    assert.Nil(t, reactor.timerQueue, "timer queue clear")
    assert.Nil(t, reactor.eventsMgr, "events mgr clear")
}

func TestReactorRegister(t *testing.T) {
    reactor, err := NewPollReactor()
    assert.Equal(t, nil, err)
    reactor.Close()
}

func TestReactorTimeout(t *testing.T) {
    reactor, err := NewPollReactor()
    assert.Nil(t, err, "create poll reactor")

    var ev TestEventHandler
    // set timer
    timerID, err := reactor.SetTimer(ev, int64(10*time.Second), nil)
    assert.Equal(t, uint64(1), timerID, "reactor set timer")
    assert.Nil(t, err, "reactor set timer")
    assert.Equal(t, uint64(2), reactor.timerID, "reactor set timer")
    assert.Equal(t, 1, reactor.timerQueue.Len(), "reactor set timer")

    // remove timer
    err = reactor.CancelTimer(timerID)
    assert.Nil(t, err, "reactor cancel timer")
    assert.Equal(t, 0, reactor.timerQueue.Len(), "reactor cancel timer")

    reactor.Close()
}
