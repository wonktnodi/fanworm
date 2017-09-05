package main

import (
    "github.com/wonktnodi/go-revolver/utils"
    log "github.com/wonktnodi/go-revolver/utils/logmate"
    "github.com/wonktnodi/go-revolver/base"
)

func main() {
    logger := log.Start(log.LogFilePath("./log"), log.EveryHour, log.AlsoStdout,
        log.LogFlags(log.Lfunc|log.Lfile|log.Lline))
    defer logger.Stop()

    if err := base.InitConfig(); err != nil {
        log.Fatalln(err)
    }

    // white list log
    //if wl, err := base.NewWhitelist(base.Conf.WhiteLog, base.Conf.Whitelist); err != nil {
    //    panic(err)
    //} else {
    //    base.DefaultWhitelist = wl
    //}

    operator := new(base.DefaultOperator)
    base.DefaultServer = base.NewServer(
        base.ServerOptions{
            CliProto:         base.Conf.CliProto,
            SvrProto:         base.Conf.SvrProto,
            HandshakeTimeout: base.Conf.HandshakeTimeout,
            TCPKeepalive:     base.Conf.TCPKeepalive,
            TCPRcvbuf:        base.Conf.TCPRcvbuf,
            TCPSndbuf:        base.Conf.TCPSndbuf,
        }, operator)

    // tcp comet
    if err := base.InitTCPTransport(base.Conf.TCPBind, base.Conf.MaxProc); err != nil {
        log.Fatalln(err)
    }

    utils.ServiceSignal()
}
