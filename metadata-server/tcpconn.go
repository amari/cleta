package main

import "net"

type TCPConn struct {
	*net.TCPConn

	hardwareAddr HardwareAddr
}

func (c TCPConn) RemoteAddr() net.Addr {
	//return c.TCPConn.RemoteAddr()
	return c.hardwareAddr
}
