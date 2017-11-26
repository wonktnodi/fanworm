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

    //reactor.EventLoop()
    reactor.Start()

    time.Sleep(time.Second)

    // try to close reactor
    reactor.Close()
    assert.Equal(t, int64(1), reactor.exitFlag, "reactor open and close")
    assert.Nil(t, reactor.events, "events clear")
    assert.Nil(t, reactor.timerQueue, "timer queue clear")
    assert.Nil(t, reactor.eventsMgr, "events mgr clear")
}

func TestReactorRegister(t *testing.T) {
    reactor, err := NewPollReactor()
    assert.Equal(t, nil, err)
    reactor.Close()
}

func TestReactorTimer(t *testing.T) {
    reactor, err := NewPollReactor()
    assert.Nil(t, err, "create poll reactor")
    timerTriggered = 0

    var ev TestEventHandler
    // set timer
    timerID, err := reactor.SetTimer(ev, int64(1*time.Second), nil)
    assert.Equal(t, uint64(1), timerID, "reactor set timer")
    assert.Nil(t, err, "reactor set timer")
    assert.Equal(t, uint64(2), reactor.timerID, "reactor set timer")
    assert.Equal(t, 1, reactor.timerQueue.Len(), "reactor set timer")

    // remove timer
    err = reactor.CancelTimer(timerID)
    assert.Nil(t, err, "reactor cancel timer")
    assert.Equal(t, 0, reactor.timerQueue.Len(), "reactor cancel timer")


    timerID, err = reactor.SetTimer(ev, int64(1 * time.Second), nil)
    assert.Equal(t, uint64(2), timerID, "reactor set timer")
    reactor.Start()

    time.Sleep(time.Second * 2)
    assert.Equal(t, 1, timerTriggered, "reactor timer timeout")

    reactor.Close()
}

func TestReactorTimeout(t *testing.T) {
    reactor, err := NewPollReactor()
    assert.Nil(t, err, "reactor timer timeout")

    var ev TestEventHandler
    timerID, err := reactor.SetTimer(ev, int64(1 * time.Second), nil)
    assert.NotEqual(t, int64(0), timerID, "reactor timer timeout")




    reactor.Close()
}