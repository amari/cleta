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

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/amari/cloud-metadata-server/internal/pkg/arp"
	models "github.com/amari/cloud-metadata-server/pkg/models/net"
)

func main() {
	arpWatcher, err := arp.NewWatcher(1 * time.Millisecond)
	if err != nil {
		log.Fatalln(err)
	}

	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 0,
		Zone: "",
	})
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(listener.Addr().String())

	srv := http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// identify the hardware address
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			remoteIP := net.ParseIP(host)
			if remoteIP == nil {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			addr := arpWatcher.GetHardwareAddrForIP4(remoteIP)
			if addr == nil {
				arpWatcher.ForcePoll()
				addr = arpWatcher.GetHardwareAddrForIP4(remoteIP)
			}
			if addr == nil {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			macAddr := models.MACAddr(addr)
			canonicalAddr := macAddr.CanonicalString()
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, r.RemoteAddr)
			fmt.Fprintln(w, macAddr.HumanReadableString())
			fmt.Fprintln(w, canonicalAddr)

			return
		}),
	}
	err = srv.Serve(listener)
	if err != nil {
		log.Fatalln(err)
	}
}
