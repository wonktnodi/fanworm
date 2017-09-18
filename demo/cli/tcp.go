package main

import (
    "time"
    "net"
    log "github.com/wonktnodi/go-revolver/utils/logmate"
    "github.com/wonktnodi/go-revolver/utils/bufio"
    "encoding/binary"
)

func initTCP() {
    conn, err := net.Dial("tcp", Conf.TCPAddr)
    if err != nil {
        log.Errorf("net.Dial(\"%s\") error(%v)", Conf.TCPAddr, err)
        return
    }
    seqId := int32(0)
    wr := bufio.NewWriter(conn)
    rd := bufio.NewReader(conn)
    proto := new(Proto)
    proto.Ver = 1
    // auth
    // test handshake timeout
    // time.Sleep(time.Second * 31)
    proto.Operation = OP_AUTH
    proto.SeqId = seqId
    proto.Body = []byte("test")
    if err = tcpWriteProto(wr, proto); err != nil {
        log.Errorf("tcpWriteProto() error(%v)", err)
        return
    }
    if err = tcpReadProto(rd, proto); err != nil {
        log.Errorf("tcpReadProto() error(%v)", err)
        return
    }
    log.Debugf("auth ok, proto: %v", proto)
    seqId++
    // writer
    go func() {
        proto1 := new(Proto)
        for {
            // heartbeat
            proto1.Operation = OP_HEARTBEAT
            proto1.SeqId = seqId
            proto1.Body = nil
            if err = tcpWriteProto(wr, proto1); err != nil {
                log.Errorf("tcpWriteProto() error(%v)", err)
                return
            }
            // test heartbeat
            //time.Sleep(time.Second * 31)
            seqId++
            // op_test
            proto1.Operation = OP_TEST
            proto1.SeqId = seqId
            if err = tcpWriteProto(wr, proto1); err != nil {
                log.Errorf("tcpWriteProto() error(%v)", err)
                return
            }
            seqId++
            time.Sleep(10000 * time.Millisecond)
        }
    }()
    // reader
    for {
        if err = tcpReadProto(rd, proto); err != nil {
            log.Errorf("tcpReadProto() error(%v)", err)
            return
        }
        if proto.Operation == OP_HEARTBEAT_REPLY {
            log.Debugf("receive heartbeat")
            if err = conn.SetReadDeadline(time.Now().Add(25 * time.Second)); err != nil {
                log.Errorf("conn.SetReadDeadline() error(%v)", err)
                return
            }
        } else if proto.Operation == OP_TEST_REPLY {
            log.Debugf("body: %s", string(proto.Body))
        } else if proto.Operation == OP_SEND_SMS_REPLY {
            log.Debugf("body: %s", string(proto.Body))
        }
    }
}

func tcpWriteProto(wr *bufio.Writer, proto *Proto) (err error) {
    // write
    if err = binary.Write(wr, binary.BigEndian, uint32(rawHeaderLen)+uint32(len(proto.Body))); err != nil {
        return
    }
    if err = binary.Write(wr, binary.BigEndian, rawHeaderLen); err != nil {
        return
    }
    if err = binary.Write(wr, binary.BigEndian, proto.Ver); err != nil {
        return
    }
    if err = binary.Write(wr, binary.BigEndian, proto.Operation); err != nil {
        return
    }
    if err = binary.Write(wr, binary.BigEndian, proto.SeqId); err != nil {
        return
    }
    if proto.Body != nil {
        log.Debugf("cipher body: %v", proto.Body)
        if err = binary.Write(wr, binary.BigEndian, proto.Body); err != nil {
            return
        }
    }
    err = wr.Flush()
    return
}

func tcpReadProto(rd *bufio.Reader, proto *Proto) (err error) {
    var (
        packLen   int32
        headerLen int16
    )
    // read
    if err = binary.Read(rd, binary.BigEndian, &packLen); err != nil {
        return
    }
    log.Debugf("packLen: %d", packLen)
    if err = binary.Read(rd, binary.BigEndian, &headerLen); err != nil {
        return
    }
    log.Debugf("headerLen: %d", headerLen)
    if err = binary.Read(rd, binary.BigEndian, &proto.Ver); err != nil {
        return
    }
    log.Debugf("ver: %d", proto.Ver)
    if err = binary.Read(rd, binary.BigEndian, &proto.Operation); err != nil {
        return
    }
    log.Debugf("operation: %d", proto.Operation)
    if err = binary.Read(rd, binary.BigEndian, &proto.SeqId); err != nil {
        return
    }
    log.Debugf("seqId: %d", proto.SeqId)
    var (
        n       = int(0)
        t       = int(0)
        bodyLen = int(packLen - int32(headerLen))
    )
    log.Debugf("read body len: %d", bodyLen)
    if bodyLen > 0 {
        proto.Body = make([]byte, bodyLen)
        for {
            if t, err = rd.Read(proto.Body[n:]); err != nil {
                return
            }
            if n += t; n == bodyLen {
                break
            } else if n < bodyLen {
            } else {
            }
        }
    } else {
        proto.Body = nil
    }
    return
}
