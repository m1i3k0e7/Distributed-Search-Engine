package net

import (
	"errors"
	"net"
)

// Get Local IP
func GetLocalIP() (ipv4 string, err error) {
	var (
		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet
		isIpNet bool
	)

	// Get all network interfaces
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}

	// Get the first non-loopback IPv4 address
	for _, addr = range addrs {
		// Check the address type and if it is not a loopback then display it
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet {
			if !ipNet.IP.IsLoopback() {
				if ipNet.IP.IsPrivate() {
					// Select IPv4, skip IPv6
					if ipNet.IP.To4() != nil {
						ipv4 = ipNet.IP.String()
						return
					}
				}
			}
		}
	}

	err = errors.New("ERR_NO_LOCAL_IP_FOUND")
	return
}
