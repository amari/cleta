package main

import (
	"C"
	"bytes"
	"encoding/binary"
	"net"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)
import (
	"io"
	"log"
)

func ioctl(fd uintptr, req uintptr, arg uintptr) (err error) {
	_, _, ep := syscall.Syscall(syscall.SYS_IOCTL, fd, req, arg)

	if ep < 0 {
		return syscall.Errno(ep)
	}
	return nil
}

func sysctl(mib []C.int, namelen uintptr, old *byte, oldlen *uintptr, new *byte, newlen uintptr) (err error) {
	_, _, ep := syscall.Syscall6(syscall.SYS___SYSCTL, uintptr(unsafe.Pointer(&mib[0])), namelen, uintptr(unsafe.Pointer(old)), uintptr(unsafe.Pointer(oldlen)), uintptr(unsafe.Pointer(new)), newlen)

	if ep < 0 {
		return syscall.Errno(ep)
	}
	return nil
}

var mib = []C.int{syscall.CTL_NET, syscall.AF_ROUTE, 0, syscall.AF_INET, 9 /* syscall.NET_RT_DUMPX_FLAGS */, syscall.RTF_LLINFO}

type rt_msghdr_ext struct {
	MsgLen           uint16
	Version          uint8
	Type             uint8
	IfpIndex         uint32
	Flags            uint32
	Reserved         uint32
	SockAddrBitmask  uint32
	Pid              int32 /* __darwin_pid_t */
	Seq              int
	Errno            int
	Use              uint32
	Inits            uint32
	Metrics          rt_metrics
	ReachabilityInfo rt_reach_info
}

func ReadRtMsghdrExt(r io.ReadSeeker, order binary.ByteOrder, v *rt_msghdr_ext) error {
	if err := binary.Read(r, order, &v.MsgLen); err != nil {
		return err
	}

	_, err := r.Seek(38, io.SeekCurrent)

	if err != nil {
		return err
	}

	if err := ReadRtMetrics(r, order, &v.Metrics); err != nil {
		return err
	}

	if err := ReadRtReachInfo(r, order, &v.ReachabilityInfo); err != nil {
		return err
	}

	return err
}

type rt_metrics struct {
	Locks                          uint32
	MTU                            uint32
	MaxExpectedHopCount            uint32
	RouteLifetime                  int32
	InboundDelayBandwidthProduct   uint32
	OutboundDelayBandwidthProduct  uint32
	OutboundGatewayBufferLimit     uint32
	EstimatedRoundTripTime         uint32
	EstimatedRoundTripTimeVariance uint32
	PacketsSent                    uint32
	Filler                         [4]uint32
}

func ReadRtMetrics(r io.ReadSeeker, order binary.ByteOrder, v *rt_metrics) error {
	_, err := r.Seek(56, io.SeekCurrent)

	return err
}

type rt_reach_info struct {
	ReferenceCount      uint32
	ProbeCount          uint32
	SndExpire           uint64
	RcvExpire           uint64
	Rssi                int32
	LinkQualityMetric   int32
	NodeProximityMetric int32
}

func ReadRtReachInfo(r io.ReadSeeker, order binary.ByteOrder, v *rt_reach_info) error {
	_, err := r.Seek(36, io.SeekCurrent)

	return err
}

type sockaddr_inarp RawSockaddrInet4Arp

type RawSockaddrInet4Arp struct {
	Len     uint8
	Family  uint8
	Port    uint16
	Addr    [4]byte
	SrcAddr [4]byte
	Tos     uint16
	Other   uint16
}

func ReadRawSockaddrInet4Arp(r io.ReadSeeker, order binary.ByteOrder, v *RawSockaddrInet4Arp) error {
	if err := binary.Read(r, order, &v.Len); err != nil {
		return err
	}

	if err := binary.Read(r, order, &v.Family); err != nil {
		return err
	}

	if err := binary.Read(r, order, &v.Port); err != nil {
		return err
	}

	if err := binary.Read(r, order, &v.Addr); err != nil {
		return err
	}

	if err := binary.Read(r, order, &v.SrcAddr); err != nil {
		return err
	}

	if err := binary.Read(r, order, &v.Tos); err != nil {
		return err
	}

	if err := binary.Read(r, order, &v.Other); err != nil {
		return err
	}

	return nil
}

// sockaddr_dl -> sys.unix.RawSockaddrDatalink
type sockaddr_dl unix.RawSockaddrDatalink

func (l *TCPListener) Accept() (net.Conn, error) {
	conn, err := l.TCPListener.AcceptTCP()

	if err != nil {
		return nil, err
	}

	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		panic("unable to split the host from the port")
		return nil, err
	}

	ipAddr := net.ParseIP(host)
	if ipAddr == nil {
		panic("unable to resolve host")
		return nil, nil
	}

	ipAddr = ipAddr.To4()

	// get raw the routing table from the kernel
	var needed uintptr
	if err := sysctl(mib, 6, nil, &needed, nil, 0); err != nil {
		panic("unable to get the routing table")
		return nil, err
	}

	// loop until the buffer is large enough to hold the routing table
	// TODO: is this a race? we may get stuck if sysctl now yields an error.
	var routingTableBuf []byte
	for {
		routingTableBuf = make([]byte, needed)

		err := sysctl(mib, 6, &routingTableBuf[0], &needed, nil, 0)

		if err == nil {
			break
		}
		needed = needed + needed/8
	}

	// parse the routing table until we find the desired ip address
	routingTableReader := bytes.NewReader(routingTableBuf)

	var (
		msgHdr       rt_msghdr_ext
		addrInet4Arp RawSockaddrInet4Arp
		addrDataLink unix.RawSockaddrDatalink
	)
	var (
		addr HardwareAddr
	)

	for {
		if routingTableReader.Len() == 0 {
			panic("routingTableReader.Len() == 0")
			//return nil, nil
		}

		offsetFromEnd := routingTableReader.Len()

		if err := ReadRtMsghdrExt(routingTableReader, binary.LittleEndian, &msgHdr); err != nil {
			panic("unable to parse the routing message header")
			return nil, err
		}

		if err := ReadRawSockaddrInet4Arp(routingTableReader, binary.LittleEndian, &addrInet4Arp); err != nil {
			panic("unable to parse the inet address")
			return nil, err
		}

		if err = binary.Read(routingTableReader, binary.LittleEndian, &addrDataLink); err != nil {
			panic("unable to parse the datalink address")
			return nil, err
		}

		diff := offsetFromEnd - routingTableReader.Len()

		routingTableReader.Seek(int64(msgHdr.MsgLen)-int64(diff), io.SeekCurrent)

		if bytes.Compare(ipAddr, addrInet4Arp.SrcAddr[:]) == 0 {
			addr = make([]byte, len(addrDataLink.Data[12-addrDataLink.Alen:][:6]))

			for i, v := range addrDataLink.Data[12-addrDataLink.Alen:][:6] {
				addr[i] = byte(v)
			}

			return &TCPConn{
				conn,
				addr,
			}, nil
		} else {
			log.Printf("%+v != %+v\n", ipAddr, addrInet4Arp.SrcAddr)
		}
	}

	// 1. skip the message header
	// 2. read the RawSockaddrInet4Arp
	// 3. read the RawSockaddrDatalink
	// 4. compare the remote address with the RawSockaddrInet4Arp.SrcAddr
	// 5. obtain the mac address at RawSockaddrDatalink.Data[RawSockaddrDatalink.NLen:]

	// rt_msghdr_ext, sockaddr_inarp, sockaddr_dl

	// search the parsed routing table

	// get the routing table

	/*
		host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err != nil {
			return nil, err
		}


			ipAddr, err := net.ResolveIPAddr("ip", host)
			if err != nil {
				return nil, err
			}*/

	// get the remote mac address from the arp tables
	// TODO: what if this address changes? we need to observe the ARP tables.

}
