package base

import (
    "net/http"
    "syscall"
    "github.com/wonktnodi/go-revolver/internal"
)

type Reactor struct {
    pollerFD   int
    timerQueue *internal.TimeoutQueue
}

func NewReactor() (*Reactor, error) {
    p, err := internal.MakePoller()
    if err != nil {
        return nil, err
    }
    var ret Reactor
    ret.pollerFD = p
    ret.timerQueue = internal.NewTimeoutQueue()

    return &ret, nil
}

func (r *Reactor) Close() {
    syscall.Close(r.pollerFD)

}

func (r *Reactor) Serve() (err error) {
    return
}

func (r *Reactor) RegisterHandler(handler EventHandler) (err error) {
    return
}

func (r *Reactor) RemoveHandler(handler http.Handler) (err error) {
    return
}
