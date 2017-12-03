package demo

import (
    "syscall"
)

type PollBase struct {
    p      int // handle for polling
    events []syscall.Kevent_t
}

func makeEvents(p *PollBase, n int) {
    p.events = make([]syscall.Kevent_t, n)
}
