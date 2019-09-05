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
package arp

import (
	"context"
	"errors"
	"net"
	"unsafe"

	"golang.org/x/sys/unix"
)

type Entry struct {
	InterfaceIndex int
	// E.g. unix.AF_INET
	Family          int
	DestinationAddr []byte
	LinkLayerAddr   []byte
	// TODO: cache statistics

	Flags Flags
	State State
}

func (e *Entry) RemoteIP() net.IP {
	if e.Family == unix.AF_INET {
		return net.IP(e.DestinationAddr)
	}

	if e.Family == unix.AF_INET6 {
		return net.IP(e.DestinationAddr)
	}

	return nil
}

func (e *Entry) HardwareAddr() net.HardwareAddr {
	return e.LinkLayerAddr
}

func (e *Entry) Interface() (*net.Interface, error) {
	return net.InterfaceByIndex(e.InterfaceIndex)
}

type Flags struct {
	IPv6Router    bool
	ProxyARPEntry bool
}

type State int

const (
	Incomplete State = unix.NUD_INCOMPLETE
	Reachable  State = unix.NUD_REACHABLE
	Stale      State = unix.NUD_STALE
	Delay      State = unix.NUD_DELAY
	Probe      State = unix.NUD_PROBE
	Failed     State = unix.NUD_FAILED
	NoARP      State = unix.NUD_NOARP
	Permanent  State = unix.NUD_PERMANENT
)

// An Table is the system defined ARP table cache.
type Table struct {
}

func NewTable() (*Table, error) {
	return &Table{}, nil
}

func (t *Table) Close() (err error) {
	return nil
}

type request struct {
	Nl unix.NlMsghdr
	Rt unix.NdMsg
}

var (
	errBadNlMsg = errors.New("Bad NlMsg")
	errNlMsgErr = errors.New("NlMsg error")
)

// Poll polls the system arp table
func (t *Table) Poll(ctx context.Context, f PollFunc) (err error) {
	//fmt.Println("polling")

	// RTM_GETNEIGH
	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_DGRAM, unix.NETLINK_ROUTE)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	err = unix.Bind(fd, &unix.SockaddrNetlink{
		Family: uint16(unix.AF_INET),
	})
	if err != nil {
		panic(err)
		return err
	}

	// send the request
	reqBytes := make([]byte, int(nlMsgLength(uintptr(unix.SizeofNlMsghdr+unix.SizeofNdMsg))))
	req := (*request)(unsafe.Pointer(&reqBytes[0]))
	req.Nl.Len = uint32(nlMsgLength(uintptr(unix.SizeofNlMsghdr + unix.SizeofNdMsg)))
	req.Nl.Flags = unix.NLM_F_REQUEST | unix.NLM_F_DUMP
	req.Nl.Type = unix.RTM_GETNEIGH
	req.Rt.State = unix.NUD_REACHABLE
	req.Rt.Family = unix.AF_INET
	req.Nl.Pid = uint32(unix.Getpid())
	req.Nl.Seq = 0
	_, err = unix.Write(fd, reqBytes)
	if err != nil {
		return err
	}
	//fmt.Printf("sent the request. %v bytes written\n", bytesWritten)

	// read the response
	buf := make([]byte, 0, 1<<13)
	tmpBuf := make([]byte, 1<<9)
	bufLen := 0

	for {
		//fmt.Println("reading a chunk")
		len, err := unix.Read(fd, tmpBuf)
		//fmt.Println("read a chunk")
		if err != nil {
			return err
		}

		nlmsg := (*unix.NlMsghdr)(unsafe.Pointer(&tmpBuf[0]))
		bufLen += len
		buf = append(buf, tmpBuf[:len]...)

		if !nlMsgOK(nlmsg, bufLen) {
			return errBadNlMsg
		}

		if nlmsg.Type == unix.NLMSG_ERROR {
			return errNlMsgErr
		}
		if nlmsg.Type == unix.NLMSG_DONE {
			break
		}
		if nlmsg.Flags&unix.NLM_F_MULTI == 0 {
			//fmt.Println("end of the message")
			break
		}
	}

	// parse the response
	nlmsg := (*unix.NlMsghdr)(unsafe.Pointer(&buf[0]))
	for {
		// parse message
		ndmsg := (*unix.NdMsg)(nlMsgData(nlmsg))

		if !nlMsgOK(nlmsg, bufLen) {
			//log.Println("no more netlink messages")
			break
		}

		//fmt.Printf("%+v\n", nlmsg)
		//fmt.Printf("%+v\n", ndmsg)

		if nlmsg.Type == unix.NLMSG_ERROR {
			return errNlMsgErr
		}

		attr := rtmRTA(unsafe.Pointer(ndmsg))
		len := rtmPayload(nlmsg)

		var entry Entry
		entry.Family = int(ndmsg.Family)
		entry.InterfaceIndex = int(ndmsg.Ifindex)

		for {
			if !rtaOK(attr, len) {
				//log.Println("no more attributes")
				break
			}
			// parse attribute
			switch attr.Type {
			case unix.NDA_UNSPEC:
			case unix.NDA_LLADDR:
				// link layer address
				// RTA_DATA(attr), RTA_PAYLOAD(attr)
				off := int(uintptr(rtaData(attr)) - uintptr(unsafe.Pointer(&buf[0])))
				addrLen := int(rtaPayload(attr))
				entry.LinkLayerAddr = buf[off : off+addrLen]
			case unix.NDA_DST:
				// destination address
				// RTA_DATA(attr), RTA_PAYLOAD(attr)
				off := int(uintptr(rtaData(attr)) - uintptr(unsafe.Pointer(&buf[0])))
				addrLen := int(rtaPayload(attr))
				entry.DestinationAddr = buf[off : off+addrLen]
			case unix.NDA_CACHEINFO:
			default:
			}
			// move to the next attribute
			attr, len = rtaNext(attr, len)
		}
		if ndmsg.Family == unix.AF_INET {
			f(ctx, entry)
		}

		nlmsg, bufLen = nlMsgNext(nlmsg, bufLen)
	}

	return nil
}

func nlMsgOK(nlh *unix.NlMsghdr, len int) bool {
	return (uintptr(len) >= uintptr(unix.SizeofNlMsghdr)) && (uintptr(nlh.Len) >= uintptr(unix.SizeofNlMsghdr)) && (uintptr(nlh.Len) <= uintptr(len))
}

func nlMsgAlign(len uintptr) uintptr {
	a := uintptr(unix.NLMSG_ALIGNTO) - 1
	b := len + uintptr(unix.NLMSG_ALIGNTO) - 1

	return b & ^a
}

func nlMsgNext(nlh *unix.NlMsghdr, len int) (*unix.NlMsghdr, int) {
	a := nlMsgAlign(uintptr(nlh.Len))
	newLen := len - int(a)
	newPtr := (*unix.NlMsghdr)(unsafe.Pointer(uintptr(unsafe.Pointer(nlh)) + nlMsgAlign(a)))

	//fmt.Printf("nlMsgNext(%p, %v) -> (%p, %v)\n", nlh, len, newPtr, newLen)

	return newPtr, newLen
}

func nlMsgLength(len uintptr) uintptr {
	return len + uintptr(unix.NLMSG_HDRLEN)
}

func nlMsgSpace(len uintptr) uintptr {
	return nlMsgAlign(nlMsgLength(len))
}

func nlMsgData(nlh *unix.NlMsghdr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(unsafe.Pointer(nlh)) + nlMsgLength(0))
}

func nlMsgPayload(nlh *unix.NlMsghdr, len uintptr) uintptr {
	return uintptr(nlh.Len) - nlMsgSpace(len)
}

func rtmRTA(r unsafe.Pointer) *unix.RtAttr {
	return (*unix.RtAttr)(unsafe.Pointer(uintptr(r) + nlMsgAlign(unix.SizeofRtMsg)))
}

func rtmPayload(nlh *unix.NlMsghdr) uintptr {
	return nlMsgPayload(nlh, uintptr(unix.SizeofRtMsg))
}

func rtaOK(rta *unix.RtAttr, len uintptr) bool {
	return (len >= uintptr(unix.SizeofRtAttr)) && (uintptr(rta.Len) >= uintptr(unix.SizeofRtAttr)) && (uintptr(rta.Len) <= len)
}

func rtaAlign(len uintptr) uintptr {
	a := uintptr(unix.RTA_ALIGNTO) - 1
	b := len + uintptr(unix.RTA_ALIGNTO) - 1

	return b & ^a
}

func rtaNext(rta *unix.RtAttr, len uintptr) (*unix.RtAttr, uintptr) {
	a := rtaAlign(uintptr(rta.Len))
	newLen := len - a
	newPtr := (*unix.RtAttr)(unsafe.Pointer(uintptr(unsafe.Pointer(rta)) + rtaAlign(a)))

	//fmt.Printf("rtaNext(%p, %v) -> (%p, %v)\n", rta, int(len), newPtr, int(newLen))

	return newPtr, newLen
}

func rtaLength(len uintptr) uintptr {
	return rtaAlign(uintptr(unix.SizeofRtAttr)) + len
}

func rtaSpace(len uintptr) uintptr {
	return rtaAlign(rtaLength(len))
}

func rtaData(rta *unix.RtAttr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(unsafe.Pointer(rta)) + rtaLength(0))
}

func rtaPayload(rta *unix.RtAttr) uintptr {
	return uintptr(rta.Len) - rtaLength(0)
}
