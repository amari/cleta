/*
Copyright © 2019 Amari Robinson

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
//go:generate stringer -type=RtMsgType

package arp

// #cgo CFLAGS: -g -Wall
// #include <stdlib.h>
// #include <net/if.h>
import "C"

import (
	"context"
	"syscall"
	"unsafe"

	"net"

	"golang.org/x/sys/unix"
)

type Entry struct {
	Header          *RtMsghdrExt
	RawIPAddr       *RawSockaddrInet4Arp
	RawDataLinkAddr *RawSockaddrDatalink
}

func (e *Entry) RemoteIP() net.IP {
	return e.RawIPAddr.RemoteIP()
}

func (e *Entry) HardwareAddr() net.HardwareAddr {
	return e.RawDataLinkAddr.HardwareAddr()
}

func (e *Entry) Interface() (*net.Interface, error) {
	return e.RawDataLinkAddr.Interface()
}

// An Table is the system defined ARP table cache.
type Table struct{}

func NewTable() (*Table, error) {
	return &Table{}, nil
}

func (a *Table) Close() (err error) {
	return nil
}

// Poll polls the system arp table
func (a *Table) Poll(ctx context.Context, f PollFunc) (err error) {
	// read the raw routing table via sysctl
	var needed uintptr
	if err := sysctl(mib, 6, nil, &needed, nil, 0); err != nil {
		return err
	}

	if needed == 0 {
		// the table is empty.
		return nil
	}

	var rawRoutingTableBuf []byte
L:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			rawRoutingTableBuf = make([]byte, needed)

			err := sysctl(mib, 6, &rawRoutingTableBuf[0], &needed, nil, 0)

			if err == nil {
				break L
			}
			//needed = needed + needed/8
			needed = needed + (needed >> 1)
		}
	}

	// parse the routing table.
	i := 0
	for i < len(rawRoutingTableBuf) {
		// cast &rawRoutingTable[i]	to *RtMsghdrExt
		msgHdr := (*RtMsghdrExt)(unsafe.Pointer(&rawRoutingTableBuf[i]))

		// sockaddr_inarp
		inArpAddr := (*RawSockaddrInet4Arp)(unsafe.Pointer(uintptr(unsafe.Pointer(msgHdr)) + unsafe.Sizeof(RtMsghdrExt{})))

		// sockaddr_dl
		dlAddr := (*RawSockaddrDatalink)((*unix.RawSockaddrDatalink)(unsafe.Pointer(uintptr(unsafe.Pointer(inArpAddr)) + uintptr(inArpAddr.Len))))

		f(ctx, Entry{msgHdr, inArpAddr, dlAddr})

		i += int(msgHdr.MsgLen)
	}

	return nil
}

// The mib for the system routing table.
var mib = []C.int{syscall.CTL_NET, syscall.AF_ROUTE, 0, syscall.AF_INET, 9 /* syscall.NET_RT_DUMPX_FLAGS */, syscall.RTF_LLINFO}

func sysctl(mib []C.int, namelen uintptr, old *byte, oldlen *uintptr, new *byte, newlen uintptr) (err error) {
	_, _, ep := syscall.Syscall6(syscall.SYS___SYSCTL, uintptr(unsafe.Pointer(&mib[0])), namelen, uintptr(unsafe.Pointer(old)), uintptr(unsafe.Pointer(oldlen)), uintptr(unsafe.Pointer(new)), newlen)

	if ep < 0 {
		return syscall.Errno(ep)
	}
	return nil
}

type RtMsgType uint8

const (
	GetExt RtMsgType = 0x15
)

type RtMsghdrExt struct {
	MsgLen           uint16
	Version          uint8
	Type             RtMsgType /*Type             uint8*/
	IfpIndex         uint32
	Flags            uint32
	Reserved         uint32
	SockAddrBitmask  uint32
	Pid              int32 /* __darwin_pid_t */
	Seq              int32 /* int */
	Errno            int32 /* int */
	Use              uint32
	Inits            uint32
	Metrics          RtMetrics
	ReachabilityInfo RtReachInfo
}

type RtMetrics struct {
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
	RouteState                     uint32
	Filler                         [3]uint32
}

type RtReachInfo struct {
	ReferenceCount      uint32
	ProbeCount          uint32
	SndExpire           uint64
	RcvExpire           uint64
	Rssi                int32
	LinkQualityMetric   int32
	NodeProximityMetric int32
}

type RawSockaddrDatalink unix.RawSockaddrDatalink

func (s *RawSockaddrDatalink) Interface() (*net.Interface, error) {
	return net.InterfaceByIndex(int(s.Index))
}

func (s *RawSockaddrDatalink) InterfaceName() string {
	/*if s.Nlen > 0 {
		nameLen := int(s.Nlen)
		//s.Data
		// s.Data -> 0..=nameLen
		var sl = struct {
			addr uintptr
			len  int
			cap  int
		}{uintptr(unsafe.Pointer(&s.Data[0])), nameLen, nameLen}
		src := *(*[]byte)(unsafe.Pointer(&sl))
		dst := make([]byte, nameLen)
		copy(dst, src)

		return string(dst)
	}

	rawName := C.malloc(255)
	defer C.free(rawName)
	//
	if res := C.if_indextoname(C.uint(s.Index), (*C.char)(rawName)); res == nil {
		return ""
	}
	name := C.GoString((*C.char)(rawName))

	return name*/
	ret, err := s.Interface()
	if err != nil {
		panic("")
	}
	return ret.Name
}

func (s *RawSockaddrDatalink) HardwareAddr() net.HardwareAddr {
	if s.Alen > 0 {
		offset := uintptr(s.Nlen)
		addrLen := int(s.Alen)

		var sl = struct {
			addr uintptr
			len  int
			cap  int
		}{uintptr(unsafe.Pointer(&s.Data[0])) + offset, addrLen, addrLen}

		src := *(*[]byte)(unsafe.Pointer(&sl))
		dst := make([]byte, addrLen)
		copy(dst, src)

		return net.HardwareAddr(dst)
	}

	return nil
}

type RawSockaddrInet4Arp struct {
	Len     uint8
	Family  uint8
	Port    uint16
	Addr    [4]byte
	SrcAddr [4]byte
	Tos     uint16
	Other   uint16
}

func (s *RawSockaddrInet4Arp) LocalIP() net.IP {
	return net.IPv4(s.SrcAddr[0], s.SrcAddr[1], s.SrcAddr[2], s.SrcAddr[3])
}

func (s *RawSockaddrInet4Arp) RemoteIP() net.IP {
	return net.IPv4(s.Addr[0], s.Addr[1], s.Addr[2], s.Addr[3])
}
