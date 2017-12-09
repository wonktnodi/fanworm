package base

import "syscall"

type PollBase struct {
    //events []syscall.Kevent_t
}

func createPoll() (p int, err error) {
    return syscall.Kqueue()
}

func closePoll(p int) (err error) {
    return syscall.Close(p)
}

func addEvents(efd int, fd int, events uint32, modify bool) (err error) {
    ev := []syscall.Kevent_t{}
    n := 0
    if events&kReadEvent == kReadEvent {
        v := syscall.Kevent_t{}
        v.Ident = uint64(fd)
        v.Flags = syscall.EV_ADD | syscall.EV_ENABLE
        v.Filter = syscall.EVFILT_READ
        ev = append(ev, v)
        n ++
    }

    if events&kWriteEvent == kWriteEvent {
        v := syscall.Kevent_t{}
        v.Ident = uint64(fd)
        v.Flags = syscall.EV_ADD | syscall.EV_ENABLE
        v.Filter = syscall.EVFILT_WRITE
        ev = append(ev, v)
        n ++
    }
    _, err = syscall.Kevent(efd, ev, nil, nil)
    return
}

func removeEvents(efd int, fd int, events uint32, modify bool) (err error) {
    ev := []syscall.Kevent_t{}
    n := 0
    if events&kReadEvent == kReadEvent {
        v := syscall.Kevent_t{}
        v.Ident = uint64(fd)
        v.Flags = syscall.EV_DELETE
        v.Filter = syscall.EVFILT_READ
        ev = append(ev, v)
        n ++
    }

    if events&kWriteEvent == kWriteEvent {
        v := syscall.Kevent_t{}
        v.Ident = uint64(fd)
        v.Flags = syscall.EV_DELETE
        v.Filter = syscall.EVFILT_WRITE
        ev = append(ev, v)
        n ++
    }
    _, err = syscall.Kevent(efd, ev, nil, nil)
    return
}

func updateEvents(efd int, fd int, events uint32, modify bool) (err error) {
    ev := []syscall.Kevent_t{}
    n := 0

    if events&kReadEvent == kReadEvent {
        v := syscall.Kevent_t{}
        v.Ident = uint64(fd)
        v.Flags = syscall.EV_ADD | syscall.EV_ENABLE | syscall.EV_CLEAR
        v.Filter = syscall.EVFILT_READ
        ev = append(ev, v)
        n ++
    } else if modify {
        v := syscall.Kevent_t{}
        v.Ident = uint64(fd)
        v.Flags = syscall.EV_DELETE
        v.Filter = syscall.EVFILT_READ
        ev = append(ev, v)
        n ++
    }

    if events&kWriteEvent == kWriteEvent {
        v := syscall.Kevent_t{}
        v.Ident = uint64(fd)
        v.Flags = syscall.EV_ADD | syscall.EV_ENABLE | syscall.EV_CLEAR
        v.Filter = syscall.EVFILT_WRITE
        ev = append(ev, v)
        n ++
    } else if modify {
        v := syscall.Kevent_t{}
        v.Ident = uint64(fd)
        v.Flags = syscall.EV_DELETE
        v.Filter = syscall.EVFILT_WRITE
        ev = append(ev, v)
        n ++
    }

    _, err = syscall.Kevent(efd, ev, nil, nil)

    return
}
