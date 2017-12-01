package base

import "strings"

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
