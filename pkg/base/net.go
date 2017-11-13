package base

import "net"

type netConn struct {
    id       int
    wake     int64
    conn     net.Conn
    detached bool
    outbuf   []byte
    err      error
}

func (c *netConn) Read(p []byte) (n int, err error) {
    return c.conn.Read(p)
}

func (c *netConn) Write(p []byte) (n int, err error) {
    if c.detached {
        if len(c.outbuf) > 0 {
            for len(c.outbuf) > 0 {
                n, err = c.conn.Write(c.outbuf)
                if n > 0 {
                    c.outbuf = c.outbuf[n:]
                }
                if err != nil {
                    return 0, err
                }
            }
            c.outbuf = nil
        }
        var tn int
        if len(p) > 0 {
            for len(p) > 0 {
                n, err = c.conn.Write(p)
                if n > 0 {
                    p = p[n:]
                    tn += n
                }
                if err != nil {
                    return tn, err
                }
            }
            p = nil
        }
        return tn, nil
    }
    return c.conn.Write(p)
}

func (c *netConn) Close() error {
    return c.conn.Close()
}
