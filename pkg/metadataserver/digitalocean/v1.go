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

package digitalocean

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amari/cloud-metadata-server/pkg/core"
	"github.com/amari/cloud-metadata-server/pkg/models/digitalocean/v1"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/amari/cloud-metadata-server/pkg/store"
	"github.com/gorilla/mux"
)

const TypeURIV1 = digitalocean.TypeURI

type EndpointV1 struct {
	*core.Server
	*httpEndpointV1

	router *mux.Router
}

// Store implements `Endpoint`
func (e *EndpointV1) Store() store.Store {
	return e.httpEndpointV1.store
}

func (e *EndpointV1) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// check that the header is set!
	addr := r.Header.Get("X-Remote-Data-Link-Addr")
	if addr == "" {
		notFoundHandlerV1(w, r)
		return
	}

	e.router.ServeHTTP(w, r)
}

func NewEndpointV1(core *core.Server, s store.Store) *EndpointV1 {
	endpoint := &httpEndpointV1{
		Server: core,
		store:  s,
	}

	router := mux.NewRouter()

	// configure router here!
	router.NotFoundHandler = http.HandlerFunc(notFoundHandlerV1)
	router.StrictSlash(false)

	router.HandleFunc("/metadata/v1.{ext:(?:json|yaml)}", endpoint.getDroplet).Methods("GET")
	router.HandleFunc("/metadata/v1/{attr:(?:id|hostname|user-data|vendor-data|public-keys)}", endpoint.getDropletAttr).Methods("GET")
	router.HandleFunc("/metadata/v1/interfaces/{type:(?:public|private)}/{id:[0-9]+}/{attr:(?:mac|type|ipv4/address|ipv4/netmask|ipv4/gateway|ipv6/address|ipv6/cidr|ipv6/gateway)}", endpoint.getNetworkInterfaceAttr).Methods("GET")
	router.HandleFunc("/metadata/v1/interfaces/{type:(?:public)}/{id:[0-9]+}/{attr:(?:anchor_ipv4/address|anchor_ipv4/netmask|anchor_ipv4/gateway)}", endpoint.getNetworkInterfaceAttr).Methods("GET")
	router.HandleFunc("/metadata/v1/floating_ip/{attr:(?:ipv4/active|ipv4/ip_address)}", endpoint.getFloatingIPAttr).Methods("GET")
	router.HandleFunc("/metadata/v1/dns/{attr:(?:nameservers)}", endpoint.getDNSAttr).Methods("GET")
	router.HandleFunc("/metadata/v1/features/{attr:(?:dhcp_enabled)}", endpoint.getFeaturesAttr).Methods("GET")

	router.HandleFunc("/metadata/v1/", endpoint.getIndex)
	router.HandleFunc("/metadata/v1/interfaces/", endpoint.getInterfaceIndex)
	router.HandleFunc("/metadata/v1/interfaces/{type:(?:public|private)}/", endpoint.getInterfaceTypeIndex)
	router.HandleFunc("/metadata/v1/interfaces/{type:(?:public|private)}/{id:[0-9]+}/", endpoint.getEnumeratedInterfaceIndex)
	router.HandleFunc("/metadata/v1/interfaces/{type:(?:public|private)}/{id:[0-9]+}/ipv4/", endpoint.getInterfaceIPv4Index)
	router.HandleFunc("/metadata/v1/interfaces/{type:(?:public|private)}/{id:[0-9]+}/ipv6/", endpoint.getInterfaceIPv6Index)
	router.HandleFunc("/metadata/v1/interfaces/{type:(?:public)}/{id:[0-9]+}/anchor_ipv4/", endpoint.getInterfaceAnchorIPv4Index)
	router.HandleFunc("/metadata/v1/floating_ip/", endpoint.getFloatingIPIndex)
	router.HandleFunc("/metadata/v1/floating_ip/ipv4/", endpoint.getFloatingIPv4Index)
	router.HandleFunc("/metadata/v1/dns/", endpoint.getDNSIndex)
	router.HandleFunc("/metadata/v1/tags/", endpoint.getTagsIndex)
	router.HandleFunc("/metadata/v1/features/", endpoint.getFeaturesIndex)

	/*router.HandleFunc("/metadata/v1", endpoint.handleIndexMovedPermanently)
	router.HandleFunc("/metadata/v1/interfaces", endpoint.handleInterfaceIndexMovedPermanently)
	router.HandleFunc("/metadata/v1/interfaces/{type:(?:public|private)}", endpoint.handleEnumeratedInterfaceIndexMovedPermanently)
	router.HandleFunc("/metadata/v1/interfaces/{type:(?:public|private)}/{id:[0-9]+}/ipv4", endpoint.handleInterfaceIPv4IndexMovedPermanently)
	router.HandleFunc("/metadata/v1/interfaces/{type:(?:public|private)}/{id:[0-9]+}/ipv6", endpoint.handleInterfaceIPv6IndexMovedPermanently)
	router.HandleFunc("/metadata/v1/interfaces/{type:(?:public)}/{id:[0-9]+}/anchor_ipv4", endpoint.handleInterfaceAnchorIPv4IndexMovedPermanently)
	router.HandleFunc("/metadata/v1/floating_ip", endpoint.handleFloatingIPIndexMovedPermanently)
	router.HandleFunc("/metadata/v1/floating_ip/ipv4", endpoint.handleFloatingIPv4IndexMovedPermanently)
	router.HandleFunc("/metadata/v1/dns", endpoint.handleDNSIndexMovedPermanently)
	router.HandleFunc("/metadata/v1/tags", endpoint.handleTagsIndexMovedPermanently)
	router.HandleFunc("/metadata/v1/features", endpoint.handleFeaturesIndexMovedPermanently)*/

	return &EndpointV1{
		core,
		endpoint,
		router,
	}
}

type httpEndpointV1 struct {
	*core.Server

	store store.Store
}

func (e *httpEndpointV1) getDropletAttr(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	switch v["attr"] {
	case "id":
		fmt.Fprint(w, droplet.ID)
	case "hostname":
		fmt.Fprint(w, droplet.Hostname)
	case "user-data":
		fmt.Fprint(w, droplet.UserData)
	case "vendor-data":
		fmt.Fprint(w, droplet.VendorData)
	case "public-keys":
		for _, v := range droplet.PublicKeys {
			fmt.Fprintln(w, v)
		}
	default:
		l.Error("bad attribute")
		notFoundHandlerV1(w, r)
		return
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) getNetworkInterfaceAttr(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	vars := mux.Vars(r)

	switch vars["type"] {
	case "public":
		interfaceID, err := strconv.ParseInt(vars["id"], 10, 32)
		if err != nil || int(interfaceID) >= len(droplet.NetworkInterfaces.PublicInterfaces) {
			notFoundHandlerV1(w, r)
			return
		}
		networkInterface := &droplet.NetworkInterfaces.PublicInterfaces[int(interfaceID)]
		switch vars["attr"] {
		case "mac":
			fmt.Fprint(w, networkInterface.Mac.HumanReadableString())
		case "type":
			fmt.Fprint(w, "public")
		case "ipv4/address":
			if networkInterface.Ipv4 == nil {
				notFoundHandlerV1(w, r)
				return
			}
			fmt.Fprint(w, networkInterface.Ipv4.Address.String())
		case "ipv4/netmask":
			if networkInterface.Ipv4 == nil {
				notFoundHandlerV1(w, r)
				return
			}
			fmt.Fprint(w, networkInterface.Ipv4.Netmask.String())
		case "ipv4/gateway":
			if networkInterface.Ipv4 == nil {
				notFoundHandlerV1(w, r)
				return
			}
			fmt.Fprint(w, networkInterface.Ipv4.Gateway.String())
		case "ipv6/address":
			if networkInterface.Ipv6 == nil {
				notFoundHandlerV1(w, r)
				return
			}
			fmt.Fprint(w, networkInterface.Ipv4.Address.String())
		case "ipv6/cidr":
			if networkInterface.Ipv6 == nil {
				notFoundHandlerV1(w, r)
				return
			}
			fmt.Fprint(w, strconv.FormatUint(uint64(networkInterface.Ipv6.Cidr), 10))
		case "ipv6/gateway":
			if networkInterface.Ipv6 == nil {
				notFoundHandlerV1(w, r)
				return
			}
			fmt.Fprint(w, networkInterface.Ipv6.Gateway.String())
		case "anchor_ipv4/address":
			if networkInterface.AnchorIpv4 == nil {
				notFoundHandlerV1(w, r)
				return
			}
			fmt.Fprint(w, networkInterface.AnchorIpv4.Address.String())
		case "anchor_ipv4/netmask":
			if networkInterface.AnchorIpv4 == nil {
				notFoundHandlerV1(w, r)
				return
			}
			fmt.Fprint(w, networkInterface.AnchorIpv4.Netmask.String())
		case "anchor_ipv4/gateway":
			if networkInterface.AnchorIpv4 == nil {
				notFoundHandlerV1(w, r)
				return
			}
			fmt.Fprint(w, networkInterface.AnchorIpv4.Gateway.String())
		default:
			notFoundHandlerV1(w, r)
			return
		}
	case "private":
		interfaceID, err := strconv.ParseInt(vars["id"], 10, 32)
		if err != nil || int(interfaceID) >= len(droplet.NetworkInterfaces.PrivateInterfaces) {
			notFoundHandlerV1(w, r)
			return
		}
		networkInterface := &droplet.NetworkInterfaces.PrivateInterfaces[int(interfaceID)]
		switch vars["attr"] {
		case "mac":
			fmt.Fprint(w, networkInterface.Mac.HumanReadableString())
		case "type":
			fmt.Fprint(w, "private")
		case "ipv4/address":
			if networkInterface.Ipv4 == nil {
				notFoundHandlerV1(w, r)
				return
			}
			fmt.Fprint(w, networkInterface.Ipv4.Address.String())
		case "ipv4/netmask":
			if networkInterface.Ipv4 == nil {
				notFoundHandlerV1(w, r)
				return
			}
			fmt.Fprint(w, networkInterface.Ipv4.Netmask.String())
		case "ipv4/gateway":
			if networkInterface.Ipv4 == nil {
				notFoundHandlerV1(w, r)
				return
			}
			fmt.Fprint(w, networkInterface.Ipv4.Gateway.String())
		case "ipv6/address":
			if networkInterface.Ipv6 == nil {
				notFoundHandlerV1(w, r)
				return
			}
			fmt.Fprint(w, networkInterface.Ipv4.Address.String())
		case "ipv6/cidr":
			if networkInterface.Ipv6 == nil {
				notFoundHandlerV1(w, r)
				return
			}
			fmt.Fprint(w, strconv.FormatUint(uint64(networkInterface.Ipv6.Cidr), 10))
		case "ipv6/gateway":
			if networkInterface.Ipv6 == nil {
				notFoundHandlerV1(w, r)
				return
			}
			fmt.Fprint(w, networkInterface.Ipv6.Gateway.String())
		default:
			notFoundHandlerV1(w, r)
			return
		}
	default:
		notFoundHandlerV1(w, r)
		return
	}
}

func (e *httpEndpointV1) getFloatingIPAttr(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	if droplet.FloatingIP == nil {
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	switch v["attr"] {
	case "ipv4/active":
		if droplet.FloatingIP.Ipv4.Active {
			fmt.Fprint(w, "true")
		} else {
			fmt.Fprint(w, "false")
		}

	case "ipv4/address":
		if droplet.FloatingIP.Ipv4.Active {
			fmt.Fprint(w, droplet.FloatingIP.Ipv4.IPAddress.String())
		} else {
			notFoundHandlerV1(w, r)
			return
		}
	default:
		l.Error("bad attribute")
		notFoundHandlerV1(w, r)
		return
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) getDNSAttr(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	if droplet.DNS == nil {
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	switch v["attr"] {
	case "nameservers":
		for _, v := range droplet.DNS.Nameservers {
			fmt.Fprintln(w, v.String())
		}
	default:
		l.Error("bad attribute")
		notFoundHandlerV1(w, r)
		return
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) getFeaturesAttr(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	switch v["attr"] {
	case "dhcp_enabled":
		if droplet.Features.DhcpEnabled {
			fmt.Fprint(w, "true")
		} else {
			fmt.Fprint(w, "false")
		}
	default:
		l.Error("bad attribute")
		notFoundHandlerV1(w, r)
		return
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) getIndex(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	fmt.Println(w, "hostname")
	fmt.Println(w, "user-data")
	fmt.Println(w, "vendor-data")
	fmt.Println(w, "public-keys")
	fmt.Println(w, "region")

	if len(droplet.NetworkInterfaces.PrivateInterfaces) > 0 || len(droplet.NetworkInterfaces.PublicInterfaces) > 0 {
		fmt.Println(w, "interfaces/")
	}

	fmt.Println(w, "dns/")
	if droplet.FloatingIP != nil {
		fmt.Println(w, "floating_ip/")
	}

	fmt.Println(w, "tags/")
	fmt.Println(w, "features/")

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) getInterfaceIndex(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	if len(droplet.NetworkInterfaces.PublicInterfaces) > 0 {
		fmt.Fprintln(w, "public/")
	}

	if len(droplet.NetworkInterfaces.PrivateInterfaces) > 0 {
		fmt.Fprintln(w, "private/")
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) getInterfaceTypeIndex(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	switch v["type"] {
	case "public":
		for i := range droplet.NetworkInterfaces.PublicInterfaces {
			fmt.Fprintf(w, "%v/\n", i)
		}
	case "private":
		for i := range droplet.NetworkInterfaces.PrivateInterfaces {
			fmt.Fprintf(w, "%v/\n", i)
		}
	default:
		notFoundHandlerV1(w, r)
		return
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) getEnumeratedInterfaceIndex(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	networkInterfaceID, err := strconv.ParseUint(v["id"], 10, 64)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}
	switch v["type"] {
	case "public":
		if int(networkInterfaceID) > len(droplet.NetworkInterfaces.PublicInterfaces) {
			notFoundHandlerV1(w, r)
			return
		}
		fmt.Fprintln(w, "mac")
		fmt.Fprintln(w, "type")

		if droplet.NetworkInterfaces.PublicInterfaces[int(networkInterfaceID)].Ipv4 != nil {
			fmt.Fprintln(w, "ipv4/")
		}

		if droplet.NetworkInterfaces.PublicInterfaces[int(networkInterfaceID)].Ipv6 != nil {
			fmt.Fprintln(w, "ipv6/")
		}

		if droplet.NetworkInterfaces.PublicInterfaces[int(networkInterfaceID)].AnchorIpv4 != nil {
			fmt.Fprintln(w, "anchor_ipv4/")
		}
	case "private":
		if int(networkInterfaceID) > len(droplet.NetworkInterfaces.PrivateInterfaces) {
			notFoundHandlerV1(w, r)
			return
		}
		fmt.Fprintln(w, "mac")
		fmt.Fprintln(w, "type")

		if droplet.NetworkInterfaces.PrivateInterfaces[int(networkInterfaceID)].Ipv4 != nil {
			fmt.Fprintln(w, "ipv4/")
		}

		if droplet.NetworkInterfaces.PrivateInterfaces[int(networkInterfaceID)].Ipv6 != nil {
			fmt.Fprintln(w, "ipv6/")
		}
	default:
		notFoundHandlerV1(w, r)
		return
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) getInterfaceIPv4Index(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	networkInterfaceID, err := strconv.ParseUint(v["id"], 10, 64)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	switch v["type"] {
	case "public":
		if int(networkInterfaceID) > len(droplet.NetworkInterfaces.PublicInterfaces) {
			notFoundHandlerV1(w, r)
			return
		}
		ip := droplet.NetworkInterfaces.PublicInterfaces[int(networkInterfaceID)].Ipv4
		if ip == nil {
			notFoundHandlerV1(w, r)
			return
		}
		fmt.Fprintln(w, "address")
		fmt.Fprintln(w, "netmask")
		fmt.Fprintln(w, "gateway")
	case "private":
		if int(networkInterfaceID) > len(droplet.NetworkInterfaces.PrivateInterfaces) {
			notFoundHandlerV1(w, r)
			return
		}
		ip := droplet.NetworkInterfaces.PrivateInterfaces[int(networkInterfaceID)].Ipv4
		if ip == nil {
			notFoundHandlerV1(w, r)
			return
		}
		fmt.Fprintln(w, "address")
		fmt.Fprintln(w, "netmask")
		fmt.Fprintln(w, "gateway")
	default:
		notFoundHandlerV1(w, r)
		return
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) getInterfaceIPv6Index(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	networkInterfaceID, err := strconv.ParseUint(v["id"], 10, 64)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	switch v["type"] {
	case "public":
		if int(networkInterfaceID) > len(droplet.NetworkInterfaces.PublicInterfaces) {
			notFoundHandlerV1(w, r)
			return
		}
		ip := droplet.NetworkInterfaces.PublicInterfaces[int(networkInterfaceID)].Ipv6
		if ip == nil {
			notFoundHandlerV1(w, r)
			return
		}
		fmt.Fprintln(w, "address")
		fmt.Fprintln(w, "netmask")
		fmt.Fprintln(w, "gateway")
	case "private":
		if int(networkInterfaceID) > len(droplet.NetworkInterfaces.PrivateInterfaces) {
			notFoundHandlerV1(w, r)
			return
		}
		ip := droplet.NetworkInterfaces.PrivateInterfaces[int(networkInterfaceID)].Ipv6
		if ip == nil {
			notFoundHandlerV1(w, r)
			return
		}
		fmt.Fprintln(w, "address")
		fmt.Fprintln(w, "netmask")
		fmt.Fprintln(w, "gateway")
	default:
		notFoundHandlerV1(w, r)
		return
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) getInterfaceAnchorIPv4Index(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	networkInterfaceID, err := strconv.ParseUint(v["id"], 10, 64)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	switch v["type"] {
	case "public":
		if int(networkInterfaceID) > len(droplet.NetworkInterfaces.PublicInterfaces) {
			notFoundHandlerV1(w, r)
			return
		}
		ip := droplet.NetworkInterfaces.PublicInterfaces[int(networkInterfaceID)].AnchorIpv4
		if ip == nil {
			notFoundHandlerV1(w, r)
			return
		}
		fmt.Fprintln(w, "address")
		fmt.Fprintln(w, "netmask")
		fmt.Fprintln(w, "gateway")
	default:
		notFoundHandlerV1(w, r)
		return
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) getFloatingIPIndex(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	if droplet.FloatingIP == nil {
		notFoundHandlerV1(w, r)
		return
	}

	fmt.Fprint(w, "ipv4/")

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) getFloatingIPv4Index(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	if droplet.FloatingIP == nil {
		notFoundHandlerV1(w, r)
		return
	}

	fmt.Fprint(w, "active")

	if droplet.FloatingIP.Ipv4.Active {
		fmt.Fprint(w, "ip_address")
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) getDNSIndex(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	_, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	fmt.Fprint(w, "nameservers")

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) getTagsIndex(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	for _, tag := range droplet.Tags {
		fmt.Fprintln(w, tag)
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) getFeaturesIndex(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	_, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	fmt.Fprintln(w, "dhcp_enabled")

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) handleIndexMovedPermanently(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	_, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	movedPermanentlyHandlerV1(w, r)
	return
}

func (e *httpEndpointV1) handleInterfaceIndexMovedPermanently(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	switch v["attr"] {
	case "dhcp_enabled":
		if droplet.Features.DhcpEnabled {
			fmt.Fprint(w, "true")
		} else {
			fmt.Fprint(w, "false")
		}
	default:
		l.Error("bad attribute")
		notFoundHandlerV1(w, r)
		return
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) handleEnumeratedInterfaceIndexMovedPermanently(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	switch v["attr"] {
	case "dhcp_enabled":
		if droplet.Features.DhcpEnabled {
			fmt.Fprint(w, "true")
		} else {
			fmt.Fprint(w, "false")
		}
	default:
		l.Error("bad attribute")
		notFoundHandlerV1(w, r)
		return
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) handleInterfaceIPv4IndexMovedPermanently(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	switch v["attr"] {
	case "dhcp_enabled":
		if droplet.Features.DhcpEnabled {
			fmt.Fprint(w, "true")
		} else {
			fmt.Fprint(w, "false")
		}
	default:
		l.Error("bad attribute")
		notFoundHandlerV1(w, r)
		return
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) handleInterfaceIPv6IndexMovedPermanently(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	switch v["attr"] {
	case "dhcp_enabled":
		if droplet.Features.DhcpEnabled {
			fmt.Fprint(w, "true")
		} else {
			fmt.Fprint(w, "false")
		}
	default:
		l.Error("bad attribute")
		notFoundHandlerV1(w, r)
		return
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) handleInterfaceAnchorIPv4IndexMovedPermanently(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	switch v["attr"] {
	case "dhcp_enabled":
		if droplet.Features.DhcpEnabled {
			fmt.Fprint(w, "true")
		} else {
			fmt.Fprint(w, "false")
		}
	default:
		l.Error("bad attribute")
		notFoundHandlerV1(w, r)
		return
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) handleFloatingIPIndexMovedPermanently(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	switch v["attr"] {
	case "dhcp_enabled":
		if droplet.Features.DhcpEnabled {
			fmt.Fprint(w, "true")
		} else {
			fmt.Fprint(w, "false")
		}
	default:
		l.Error("bad attribute")
		notFoundHandlerV1(w, r)
		return
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) handleDNSIndexMovedPermanently(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	switch v["attr"] {
	case "dhcp_enabled":
		if droplet.Features.DhcpEnabled {
			fmt.Fprint(w, "true")
		} else {
			fmt.Fprint(w, "false")
		}
	default:
		l.Error("bad attribute")
		notFoundHandlerV1(w, r)
		return
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) handleTagsIndexMovedPermanently(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	switch v["attr"] {
	case "dhcp_enabled":
		if droplet.Features.DhcpEnabled {
			fmt.Fprint(w, "true")
		} else {
			fmt.Fprint(w, "false")
		}
	default:
		l.Error("bad attribute")
		notFoundHandlerV1(w, r)
		return
	}

	l.Info("", zap.Int("status", 200))
}

func (e *httpEndpointV1) handleFeaturesIndexMovedPermanently(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")
	l := e.Log().With(
		zap.String("datalink_addr", dataLinkAddr),
		zap.String("request_path", r.URL.String()),
		zap.String("schema", TypeURIV1),
	)

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	droplet, ok := d.Contents.(*digitalocean.Droplet)
	if !ok {
		l.Error("document not found")
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	switch v["attr"] {
	case "dhcp_enabled":
		if droplet.Features.DhcpEnabled {
			fmt.Fprint(w, "true")
		} else {
			fmt.Fprint(w, "false")
		}
	default:
		l.Error("bad attribute")
		notFoundHandlerV1(w, r)
		return
	}

	l.Info("", zap.Int("status", 200))
}

func movedPermanentlyHandlerV1(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "Moved Permanently.")
}

func notFoundHandlerV1(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMovedPermanently)
	fmt.Fprint(w, "not found")
}

func indexHandlerV1(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`id
hostname
user-data
vendor-data
public-keys
region
interfaces/
dns/
floating_ip/
tags/
features/`))
}

func (e *httpEndpointV1) getDroplet(w http.ResponseWriter, r *http.Request) {
	dataLinkAddr := r.Header.Get("X-Remote-Data-Link-Addr")

	d, err := e.store.GetDocument(r.Context(), dataLinkAddr, TypeURIV1)
	if err != nil {
		notFoundHandlerV1(w, r)
		return
	}

	v := mux.Vars(r)
	switch v["ext"] {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(d.Contents)
		if err != nil {
			e.Server.Log().Error("failed to serialize JSON", zap.String("dataLinkAddr", dataLinkAddr))
		}
	case "yaml":
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		err = yaml.NewEncoder(w).Encode(d.Contents)
		if err != nil {
			e.Server.Log().Error("failed to serialize YAML", zap.String("dataLinkAddr", dataLinkAddr))
		}
	default:
		notFoundHandlerV1(w, r)
		return
	}
}
