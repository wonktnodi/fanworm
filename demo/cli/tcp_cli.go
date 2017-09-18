package main

import (
    "github.com/abiosoft/ishell"
    log "github.com/wonktnodi/go-revolver/utils/logmate"
    "strings"
    "flag"
)

func main() {
    logger := log.Start(log.LogFilePath("./log"), log.EveryHour, log.AlsoStdout,
        log.LogFlags(log.Lfunc|log.Lfile|log.Lline))
    defer logger.Stop()

    flag.Parse()
    if err := InitConfig(); err != nil {
        panic(err)
    }

    shell := ishell.New()

    // display welcome info.
    shell.Println("Sample Interactive Shell")
    // register a function for "greet" command.
    shell.AddCmd(&ishell.Cmd{
        Name: "greet",
        Help: "greet user",
        Func: func(c *ishell.Context) {
            c.Println("Hello", strings.Join(c.Args, " "))
        },
    })

    if Conf.Type == ProtoTCP {
        initTCP()
    }

    // run shell
    shell.Start()

    return
}
