package utils

import (
    "os"
    "os/signal"
    "syscall"
    log "github.com/wonktnodi/go-revolver/utils/logmate"
    "time"
)

func ServiceSignal() {
    c := make(chan os.Signal, 1)
    signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP)
    for {
        s := <-c
        log.Infof("got a signal %s", s.String())
        switch s {
        case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
            Shutdown()
            return
        case syscall.SIGUSR2:
            Reboot()
        case syscall.SIGHUP:
        default:
            return
        }
    }
}

func Shutdown(timeout ...time.Duration) {
    log.Infof("shutting down process...")
}

func Reboot(timeout ...time.Duration) {
    log.Infof("rebooting process...")
}

