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

package metadataserver

import (
	"net"
	"net/http"
	"time"

	"github.com/amari/cloud-metadata-server/internal/pkg/arp"
	"github.com/amari/cloud-metadata-server/pkg/core"
	model "github.com/amari/cloud-metadata-server/pkg/models/net"
	"github.com/amari/cloud-metadata-server/pkg/store"
	"go.uber.org/zap"
)

type HTTPServer struct {
	*core.Server

	arpWatcher *arp.Watcher
	router     *Router
	store      store.Store
}

func NewHTTPServer(c *core.Server, s store.Store, d time.Duration) (*HTTPServer, error) {
	w, err := arp.NewWatcher(d)
	if err != nil {
		return nil, err
	}

	r := NewRouter(c, s)

	return &HTTPServer{
		Server:     c.WithLoggerFields(zap.String("endpoint", "http")),
		arpWatcher: w,
		router:     r,
		store:      s,
	}, nil
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	addr := s.arpWatcher.GetHardwareAddrForIP4(remoteIP)
	if addr == nil {
		s.arpWatcher.ForcePoll()
		addr = s.arpWatcher.GetHardwareAddrForIP4(remoteIP)
	}
	if addr == nil {
		s.Log().Error("data link addr not found", zap.String("remoteAddr", r.RemoteAddr))
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	canonicalAddr := model.MACAddr(addr).CanonicalString()
	r.Header.Set("X-Remote-Data-Link-Addr", canonicalAddr)
	// identify the type uri and serve the request
	typeURIs, err := s.store.ListSupportedTypeURIs(r.Context(), canonicalAddr)
	if err != nil {
		s.Log().Error("typeURI not found", zap.String("remoteAddr", r.RemoteAddr), zap.String("canonicalRemoteDataLinkAddr", canonicalAddr))
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	for _, typeURI := range typeURIs {
		if endpoint, ok := s.router.Match(typeURI).(HTTPEndpoint); ok && endpoint != nil {
			endpoint.ServeHTTP(w, r)
			return
		}
	}
	//
	http.Error(w, "", http.StatusNotFound)

	return
}
