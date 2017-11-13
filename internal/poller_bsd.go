package internal

import (
    "syscall"
    "time"
)

func MakePoller() (p int, err error) {
    return syscall.Kqueue()
}

func MakeEvents(n int) interface{} {
    return make([]syscall.Kevent_t, n)
}

func Wait(p int, evs interface{}, timeout time.Duration) (n int, err error) {
    if timeout < 0 {
        timeout = 0
    }
    ts := syscall.NsecToTimespec(int64(timeout))
    return syscall.Kevent(p, nil, evs.([]syscall.Kevent_t), &ts)
}