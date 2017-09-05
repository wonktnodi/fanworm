package base

import (
    "net"
    log "github.com/wonktnodi/go-revolver/utils/logmate"
    "github.com/wonktnodi/go-revolver/utils/timer"
    "github.com/wonktnodi/go-revolver/utils/bytes"
    "time"
    "github.com/wonktnodi/go-revolver/define"
    "github.com/wonktnodi/go-revolver/proto"
    "github.com/wonktnodi/go-revolver/utils/bufio"
    "io"
    "github.com/satori/go.uuid"
    "fmt"
)

// InitTCP listen all tcp.bind and start accept connections.
func InitTCPTransport(addrs []string, accept int) (err error) {
    var (
        bind     string
        listener *net.TCPListener
        addr     *net.TCPAddr
    )
    for _, bind = range addrs {
        if addr, err = net.ResolveTCPAddr("tcp4", bind); err != nil {
            log.Errorf("net.ResolveTCPAddr(\"tcp4\", \"%s\") error(%v)", bind, err)
            return
        }

        if listener, err = net.ListenTCP("tcp4", addr); err != nil {
            log.Errorf("net.ListenTCP(\"tcp4\", \"%s\") error(%v)", bind, err)
            return
        }

        log.Tracef("start tcp listen: \"%s\"", bind)

        // split N core accept
        for i := 0; i < accept; i++ {
            go acceptTCP(DefaultServer, listener)
        }
    }
    return
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.  Accept blocks; the caller typically
// invokes it in a go statement.
func acceptTCP(server *Server, lis *net.TCPListener) {
    var (
        conn *net.TCPConn
        err  error
        r    int
    )
    for {
        if conn, err = lis.AcceptTCP(); err != nil {
            // if listener close then return
            log.Errorf("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
            return
        }
        log.Debugf("conn[%v] incoming", conn.RemoteAddr())
        if err = conn.SetKeepAlive(server.Options.TCPKeepalive); err != nil {
            log.Errorf("conn.SetKeepAlive() error(%v)", err)
            return
        }
        if err = conn.SetReadBuffer(server.Options.TCPRcvbuf); err != nil {
            log.Errorf("conn.SetReadBuffer() error(%v)", err)
            return
        }
        if err = conn.SetWriteBuffer(server.Options.TCPSndbuf); err != nil {
            log.Errorf("conn.SetWriteBuffer() error(%v)", err)
            return
        }
        go serveTCP(server, conn, r)
        if r++; r == maxInt {
            r = 0
        }
    }
}

func serveTCP(server *Server, conn *net.TCPConn, r int) {
    var (
        // timer
        tr = server.round.Timer(r)
        rp = server.round.Reader(r)
        wp = server.round.Writer(r)
        // ip addr
        lAddr = conn.LocalAddr().String()
        rAddr = conn.RemoteAddr().String()
    )
    log.Tracef("start tcp serve \"%s\" with \"%s\"", lAddr, rAddr)

    server.serveTCP(conn, rp, wp, tr)
}

func (server *Server) serveTCP(conn *net.TCPConn, rp, wp *bytes.Pool, tr *timer.Timer) {
    var (
        err   error
        key   string
        white bool
        hb    time.Duration // heartbeat
        p     *proto.Proto
        b     *Bucket
        trd   *timer.TimerData
        rb    = rp.Get()
        wb    = wp.Get()
        ch    = NewChannel(server.Options.CliProto, server.Options.SvrProto, define.NoRoom)
        rr    = &ch.Reader
        wr    = &ch.Writer
    )

    ch.Reader.ResetBuffer(conn, rb.Bytes())
    ch.Writer.ResetBuffer(conn, wb.Bytes())

    // hanshake ok start dispatch goroutine

    // must not setadv, only used in auth
    if p, err = ch.CliProto.Set(); err == nil {
        if key, hb, err = server.authTCP(rr, wr, p); err == nil {
            b = server.Bucket(key)
            err = b.Put(key, ch)
        }
    }
    if err != nil {
        conn.Close()
        rp.Put(rb)
        wp.Put(wb)
        tr.Del(trd)
        log.Errorf("key: %s handshake failed error(%v)", key, err)
        return
    }
    //trd.Key = key
    //tr.Set(trd, hb)

    //go server.dispatchTCP(key, conn, wr, wp, wb, ch)
    for {
        if p, err = ch.CliProto.Set(); err != nil {
            break
        }
        if white {
            DefaultWhitelist.Log.Printf("key: %s start read proto\n", key)
        }
        if err = p.ReadTCP(rr); err != nil {
            break
        }
        if white {
            DefaultWhitelist.Log.Printf("key: %s read proto:%v\n", key, p)
        }
        if p.Operation == define.OP_HEARTBEAT {
            tr.Set(trd, hb)
            p.Body = nil
            p.Operation = define.OP_HEARTBEAT_REPLY

            log.Tracef("key: %s receive heartbeat", key)

        } else {
            if err = server.operator.Operate(p); err != nil {
                break
            }
        }
        if white {
            DefaultWhitelist.Log.Printf("key: %s process proto:%v\n", key, p)
        }
        ch.CliProto.SetAdv()
        ch.Signal()
        if white {
            DefaultWhitelist.Log.Printf("key: %s signal\n", key)
        }
    }
    if white {
        DefaultWhitelist.Log.Printf("key: %s server tcp error(%v)\n", key, err)
    }
    if err != nil && err != io.EOF {
        log.Errorf("key: %s server tcp failed error(%v)", key, err)
    }
    b.Del(key)
    //tr.Del(trd)
    rp.Put(rb)
    conn.Close()
    ch.Close()
    if err = server.operator.Disconnect(key, ch.RoomId); err != nil {
        log.Errorf("key: %s operator do disconnect error(%v)", key, err)
    }
    if white {
        DefaultWhitelist.Log.Printf("key: %s disconnect error(%v)\n", key, err)
    }

    log.Tracef("key: %s server tcp goroutine exit", key)

    return
}

// auth for goim handshake with client, use rsa & aes.
func (server *Server) authTCP(rr *bufio.Reader, wr *bufio.Writer, p *proto.Proto) (key string, heartbeat time.Duration, err error) {

    //if err = p.ReadTCP(rr); err != nil {
    //    return
    //}
    //if p.Operation != define.OP_AUTH {
    //    log.Warnf("auth operation not valid: %d", p.Operation)
    //    err = ErrOperation
    //    return
    //}
    //if key, rid, heartbeat, err = server.operator.Connect(p); err != nil {
    //    return
    //}
    //p.Body = nil
    //p.Operation = define.OP_AUTH_REPLY
    //if err = p.WriteTCP(wr); err != nil {
    //    return
    //}
    //err = wr.Flush()
    key = fmt.Sprint(uuid.NewV4())
    heartbeat = time.Second * 10
    return
}
