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