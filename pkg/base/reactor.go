package base

import (
    "sync"
    "github.com/wonktnodi/go-revolver/pkg/log"
)

const (
    kReadEvent  = 1;
    kWriteEvent = 2;
)

type Reactor interface {
    AddEventHandler(handler EventHandler, masks uint32) (err error)
    RemoveEventHandler(handler EventHandler, masks uint32) (err error)
    DeleteEventHandler(handler EventHandler, masks uint32) (err error)
    Open() (err error)
    Close()
    EventLoop()
    SetTimer(ev EventHandler, delay int64, data interface{}) (id uint64, err error)
    CancelTimer(id uint64) (err error)
}

var NewReactor func() (Reactor, error)

var instance Reactor
var once sync.Once

func ReactorInstance() Reactor {
    once.Do(func() {
        var err error
        instance, err = NewReactor()
        if nil != err {
            log.Fatalln("failed to create reactor instance, ", err)
        }
    })
    return instance
}

