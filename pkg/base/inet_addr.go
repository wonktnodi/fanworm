package base

import (
    "strings"
    "syscall"
    "net"
)

func parseAddr(addr string) (network, address string, stdlib bool) {
    network = "tcp"
    address = addr
    if strings.Contains(address, "://") {
        network = strings.Split(address, "://")[0]
        address = strings.Split(address, "://")[1]
    }
    if strings.HasSuffix(network, "-net") {
        stdlib = true
        network = network[:len(network)-4]
    }
    return
}

func sockaddrToAddr(sa syscall.Sockaddr) net.Addr {
    var a net.Addr
    switch sa := sa.(type) {
    case *syscall.SockaddrInet4:
        a = &net.TCPAddr{
            IP:   append([]byte{}, sa.Addr[:]...),
            Port: sa.Port,
        }
    case *syscall.SockaddrInet6:
        var zone string
        if sa.ZoneId != 0 {
            if ifi, err := net.InterfaceByIndex(int(sa.ZoneId)); err == nil {
                zone = ifi.Name
            }
        }
        if zone == "" && sa.ZoneId != 0 {
        }
        a = &net.TCPAddr{
            IP:   append([]byte{}, sa.Addr[:]...),
            Port: sa.Port,
            Zone: zone,
        }
    case *syscall.SockaddrUnix:
        a = &net.UnixAddr{Net: "unix", Name: sa.Name}
    }
    return a
}

// resolve resolves an ip address and returns a sockaddr for socket
// connection to external servers.
func resolve(addr string) (sa syscall.Sockaddr, err error) {
    network, address, _ := parseAddr(addr)
    var taddr net.Addr
    switch network {
    default:
        return nil, net.UnknownNetworkError(network)
    case "unix":
        taddr = &net.UnixAddr{Net: "unix", Name: address}
    case "tcp", "tcp4", "tcp6":
        // use the stdlib resolver because it's good.
        taddr, err = net.ResolveTCPAddr(network, address)
        if err != nil {
            return nil, err
        }
    }
    switch taddr := taddr.(type) {
    case *net.UnixAddr:
        sa = &syscall.SockaddrUnix{Name: taddr.Name}
    case *net.TCPAddr:
        switch len(taddr.IP) {
        case 0:
            var sa4 syscall.SockaddrInet4
            sa4.Port = taddr.Port
            sa = &sa4
        case 4:
            var sa4 syscall.SockaddrInet4
            copy(sa4.Addr[:], taddr.IP[:])
            sa4.Port = taddr.Port
            sa = &sa4
        case 16:
            var sa6 syscall.SockaddrInet6
            copy(sa6.Addr[:], taddr.IP[:])
            sa6.Port = taddr.Port
            sa = &sa6
        }
    }
    return sa, nil
}

func filladdrs(c *Connection) {
    if c.laddr == nil {
        sa, _ := syscall.Getsockname(c.fd)
        c.laddr = sockaddrToAddr(sa)
    }
    if c.raddr == nil {
        sa, _ := syscall.Getsockname(c.fd)
        c.raddr = sockaddrToAddr(sa)
    }
}
