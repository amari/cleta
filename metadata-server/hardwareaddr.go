package main

import "net"

type HardwareAddr net.HardwareAddr

func (a HardwareAddr) Network() string {
	return "mac"
}

func (a HardwareAddr) String() string {
	return net.HardwareAddr(a).String()
}
