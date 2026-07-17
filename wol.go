package main

import (
	"net"
	"strconv"

	"github.com/mdlayher/wol"
)

// sendMagicPacket broadcasts a Wake-on-LAN magic packet to the given MAC via
// the broadcast address and UDP port.
func sendMagicPacket(mac net.HardwareAddr, broadcast string, port int) error {
	c, err := wol.NewClient()
	if err != nil {
		return err
	}
	defer c.Close()

	addr := net.JoinHostPort(broadcast, strconv.Itoa(port))
	return c.Wake(addr, mac)
}
