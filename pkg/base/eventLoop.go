package base

import (
    "github.com/wonktnodi/go-revolver/internal"
    "syscall"
    "sync"
    "time"
)

func serve(handler EventHandler) error {
    p, err := internal.MakePoller()
    if err != nil {
        return err
    }
    defer syscall.Close(p)

    var mu sync.Mutex

    timeoutqueue := internal.NewTimeoutQueue()
    var id int

    var evs = internal.MakeEvents(64)
    nextTicker := time.Now()
    for {
        delay := nextTicker.Sub(time.Now())
        if delay < 0 {
            delay = 0
        } else if delay > time.Second/4 {
            delay = time.Second / 4
        }

        pn, err := internal.Wait(p, evs, delay)
        if err != nil && err != syscall.EINTR {
            return err
        }

        if timeoutqueue.Len() > 0 {
            var count int
            now := time.Now()
        }

    }
}
