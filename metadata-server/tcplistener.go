package main

import "net"

type TCPListener struct {
	*net.TCPListener
}

func ListenTCP(network string, address *net.TCPAddr) (*TCPListener, error) {
	listener, err := net.ListenTCP(network, address)
	if err != nil {
		return nil, err
	}

	return &TCPListener{
		listener,
	}, nil
}
