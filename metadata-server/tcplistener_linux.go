package main

import (
	"net"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

func ioctl(fd uintptr, req uintptr, arg uintptr) (err error) {
	_, _, ep := syscall.Syscall(syscall.SYS_IOCTL, fd, req, arg)
	if ep != 0 {
		return syscall.Errno(ep)
	}
	return nil
}

type arpreq struct {
	ProtocolAddress syscall.RawSockaddr
	HardwareAddress syscall.RawSockaddr
	Flags           int
	Netmask         syscall.RawSockaddr
	Dev             [16]int8
}

func (l *TCPListener) Accept() (net.Conn, error) {
	conn, err := l.TCPListener.AcceptTCP()

	if err != nil {
		return nil, err
	}

	// get the remote mac address from the arp tables
	// TODO: what if this address changes? we need to observe the ARP tables.

	rawConn, err := conn.SyscallConn()
	if err != nil {
		return nil, err
	}

	var (
		addr     HardwareAddr
		ioctlErr error
	)

	if err := rawConn.Control(func(fd uintptr) {
		var arpRequest arpreq

		ioctlErr = ioctl(fd, unix.SIOCGARP, uintptr(unsafe.Pointer(&arpRequest)))
		if ioctlErr != nil {
			return
		}

		addr = make([]byte, len(arpRequest.HardwareAddress.Data))

		for i, v := range arpRequest.HardwareAddress.Data {
			addr[i] = byte(v)
		}
	}); err != nil {
		return nil, err
	}

	if ioctlErr != nil {
		return nil, ioctlErr
	}

	return &TCPConn{
		conn,
		addr,
	}, nil
}
