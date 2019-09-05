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
	"net"
	"strconv"

	model "github.com/amari/cloud-metadata-server/pkg/models/net"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

const TypeURI = "digitalocean.com/v1"

type Droplet struct {
	ID                uint64            `json:"droplet_id" yaml:"droplet_id" toml:"droplet_id"`
	Hostname          string            `json:"hostname" yaml:"hostname" toml:"hostname"`
	UserData          UserData          `json:"user_data"  yaml:"user_data" toml:"user_data"`
	VendorData        VendorData        `json:"vendor_data"  yaml:"vendor_data" toml:"vendor_data"`
	PublicKeys        []PublicKey       `json:"public_keys"  yaml:"public_keys" toml:"public_keys"`
	Region            string            `json:"region"  yaml:"region" toml:"region"`
	NetworkInterfaces NetworkInterfaces `json:"interfaces" yaml:"interfaces" toml:"interfaces"`
	FloatingIP        *FloatingIp       `json:"floating_ip"  yaml:"floating_ip" toml:"floating_ip"`
	DNS               *DNS              `json:"dns"  yaml:"dns" toml:"dns"`
	Tags              []string          `json:"tags,omitempty"  yaml:"tags,omitempty" toml:"tags,omitempty"`
	Features          Features          `json:"features"  yaml:"features" toml:"features"`
}

func (d *Droplet) marshalMap() (map[string]interface{}, error) {
	var err error
	m := map[string]interface{}{}
	m["droplet_id"] = d.ID
	m["hostname"] = d.ID
	m["user_data"] = string(d.UserData)
	m["vendor_data"] = string(d.VendorData)
	m["region"] = d.Region
	m["interfaces"], err = d.NetworkInterfaces.marshalMap()
	if err != nil {
		return nil, err
	}
	m["floating_ip"], err = d.FloatingIP.marshalMap()
	if err != nil {
		return nil, err
	}
	m["dns"], err = d.DNS.marshalMap()
	if err != nil {
		return nil, err
	}
	m["tags"] = d.Tags
	m["features"], err = d.Features.marshalMap()
	if err != nil {
		return nil, err
	}
	/*jsonBytes, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(jsonBytes, &m)
	if err != nil {
		return nil, err
	}*/
	return m, nil
}

/*func (d *Droplet) MarshalTOML() (data []byte, err error) {
	m, err := d.marshalMap()
	if err != nil {
		return nil, err
	}

	tree, err := toml.TreeFromMap(m)
	if err != nil {
		return nil, err
	}
	s, err := tree.ToTomlString()
	return []byte(s), err
}*/

var availableDropletKeys = [...]string{"id", "hostname", "user-data", "vendor-data", "public-keys", "region", "interfaces/", "dns/", "floating_ip/", "tags/", "features/"}

/*func (d *Droplet) UnmarshalYAML(value *yaml.Node) error {
	err := value.Decode(&d.ID)
	if err != nil {
		return err
	}

	return nil
}*/

func (d *Droplet) availableKeys() []string {
	return availableDropletKeys[:]
}

func (d *Droplet) TypeURI() string {
	return TypeURI
}

func (d *Droplet) DataLinkAddrs() []model.DataLinkAddr {
	res := make([]model.DataLinkAddr, 0, len(d.NetworkInterfaces.PrivateInterfaces)+len(d.NetworkInterfaces.PublicInterfaces))

	for i := 0; i < len(d.NetworkInterfaces.PrivateInterfaces); i++ {
		res = append(res, d.NetworkInterfaces.PrivateInterfaces[i].Mac)
	}

	for i := 0; i < len(d.NetworkInterfaces.PublicInterfaces); i++ {
		res = append(res, d.NetworkInterfaces.PublicInterfaces[i].Mac)
	}

	return res
}

type UserData string
type VendorData string
type PublicKey string

type NetworkInterfaces struct {
	PrivateInterfaces []PrivateNetworkInterface `json:"private" yaml:"private" toml:"private"`
	PublicInterfaces  []PublicNetworkInterface  `json:"public" yaml:"public" toml:"public"`
}

func (d *NetworkInterfaces) marshalMap() (map[string]interface{}, error) {
	m := map[string]interface{}{}
	privateInterfaces := []interface{}{}
	for _, privateInterface := range d.PrivateInterfaces {
		m, err := privateInterface.marshalMap()
		if err != nil {
			return nil, err
		}
		privateInterfaces = append(privateInterfaces, m)
	}
	publicInterfaces := []interface{}{}
	for _, publicInterface := range d.PublicInterfaces {
		m, err := publicInterface.marshalMap()
		if err != nil {
			return nil, err
		}
		publicInterfaces = append(publicInterfaces, m)
	}
	m["private"] = privateInterfaces
	m["public"] = publicInterfaces

	return m, nil
}

type PublicNetworkInterface struct {
	Mac        model.MACAddr `json:"mac" yaml:"mac" toml:"mac"`
	Ipv4       *IPv4Addr     `json:"ipv4" yaml:"ipv4" toml:"ipv4"`
	Ipv6       *IPv6Addr     `json:"ipv6" yaml:"ipv6" toml:"ipv6"`
	AnchorIpv4 *IPv4Addr     `json:"anchor_ipv4,omitempty"  yaml:"anchor_ipv4,omitempty" toml:"anchor_ipv4,omitempty"`
}

func (d *PublicNetworkInterface) marshalMap() (map[string]interface{}, error) {
	var err error
	m := map[string]interface{}{}
	m["type"] = "public"
	m["mac"] = d.Mac.HumanReadableString()
	if d.Ipv4 != nil {
		m["ipv4"], err = d.Ipv4.marshalMap()
		if err != nil {
			return nil, err
		}
	}
	if d.Ipv6 != nil {
		m["ipv6"], err = d.Ipv6.marshalMap()
		if err != nil {
			return nil, err
		}
	}
	if d.AnchorIpv4 != nil {
		m["anchor_ipv4"], err = d.AnchorIpv4.marshalMap()
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

// MarshalJSON implements `json.Marshaler`
func (p *PublicNetworkInterface) MarshalJSON() ([]byte, error) {
	ret := struct {
		Mac        model.MACAddr `json:"mac" yaml:"mac" toml:"mac"`
		Ipv4       *IPv4Addr     `json:"ipv4" yaml:"ipv4" toml:"ipv4"`
		Ipv6       *IPv6Addr     `json:"ipv6" yaml:"ipv6" toml:"ipv6"`
		AnchorIpv4 *IPv4Addr     `json:"anchor_ipv4,omitempty"  yaml:"anchor_ipv4,omitempty" toml:"anchor_ipv4,omitempty"`
		Type       string        `json:"type"  yaml:"type" toml:"type"`
	}{
		p.Mac,
		p.Ipv4,
		p.Ipv6,
		p.AnchorIpv4,
		"public",
	}
	return json.Marshal(&ret)
}

// MarshalYAML implements `yaml.Marshaler`
func (p PublicNetworkInterface) MarshalYAML() (interface{}, error) {
	ret := struct {
		Mac        model.MACAddr `json:"mac" yaml:"mac" toml:"mac"`
		Ipv4       *IPv4Addr     `json:"ipv4" yaml:"ipv4" toml:"ipv4"`
		Ipv6       *IPv6Addr     `json:"ipv6" yaml:"ipv6" toml:"ipv6"`
		AnchorIpv4 *IPv4Addr     `json:"anchor_ipv4,omitempty"  yaml:"anchor_ipv4,omitempty" toml:"anchor_ipv4,omitempty"`
		Type       string        `json:"type"  yaml:"type" toml:"type"`
	}{
		p.Mac,
		p.Ipv4,
		p.Ipv6,
		p.AnchorIpv4,
		"public",
	}
	return &ret, nil
}

type PrivateNetworkInterface struct {
	Mac  model.MACAddr `json:"mac" yaml:"mac" toml:"mac"`
	Ipv4 *IPv4Addr     `json:"ipv4,omitempty" yaml:"ipv4,omitempty" toml:"ipv4,omitempty"`
	Ipv6 *IPv6Addr     `json:"ipv6,omitempty" yaml:"ipv6,omitempty" toml:"ipv6,omitempty"`
}

func (d *PrivateNetworkInterface) marshalMap() (map[string]interface{}, error) {
	var err error
	m := map[string]interface{}{}
	m["type"] = "private"
	m["mac"] = d.Mac.HumanReadableString()
	if d.Ipv4 != nil {
		m["ipv4"], err = d.Ipv4.marshalMap()
		if err != nil {
			return nil, err
		}
	}
	if d.Ipv6 != nil {
		m["ipv6"], err = d.Ipv6.marshalMap()
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

// MarshalJSON implements `json.Marshaler`
func (p *PrivateNetworkInterface) MarshalJSON() ([]byte, error) {
	ret := struct {
		Mac  model.MACAddr `json:"mac" yaml:"mac" toml:"mac"`
		Ipv4 *IPv4Addr     `json:"ipv4,omitempty" yaml:"ipv4,omitempty" toml:"ipv4,omitempty"`
		Ipv6 *IPv6Addr     `json:"ipv6,omitempty" yaml:"ipv6,omitempty" toml:"ipv6,omitempty"`
		Type string        `json:"type"  yaml:"type" toml:"type"`
	}{
		p.Mac,
		p.Ipv4,
		p.Ipv6,
		"private",
	}
	return json.Marshal(&ret)
}

// MarshalYAML implements `yaml.Marshaler`
func (p PrivateNetworkInterface) MarshalYAML() (interface{}, error) {
	ret := struct {
		Mac  model.MACAddr `json:"mac" yaml:"mac" toml:"mac"`
		Ipv4 *IPv4Addr     `json:"ipv4,omitempty" yaml:"ipv4,omitempty" toml:"ipv4,omitempty"`
		Ipv6 *IPv6Addr     `json:"ipv6,omitempty" yaml:"ipv6,omitempty" toml:"ipv6,omitempty"`
		Type string        `json:"type"  yaml:"type" toml:"type"`
	}{
		p.Mac,
		p.Ipv4,
		p.Ipv6,
		"private",
	}
	return &ret, nil
}

type IPv4Addr struct {
	Address model.IPv4     `json:"ip_address" yaml:"ip_address" toml:"ip_address"`
	Netmask model.IPv4Mask `json:"netmask" yaml:"netmask" toml:"netmask"`
	Gateway model.IPv4     `json:"gateway" yaml:"gateway" toml:"gateway"`
}

func (d *IPv4Addr) marshalMap() (map[string]interface{}, error) {
	m := map[string]interface{}{
		"ip_address": d.Address,
		"netmask":    d.Netmask,
		"gateway":    d.Gateway,
	}

	return m, nil
}

type IPv6Addr struct {
	Address model.IPv6 `json:"ip_address" yaml:"ip_address" toml:"ip_address"`
	// CIDR block size
	Cidr    uint8      `json:"cidr" yaml:"cidr" toml:"cidr"`
	Gateway model.IPv6 `json:"gateway" yaml:"gateway" toml:"gateway"`
}

func (d *IPv6Addr) marshalMap() (map[string]interface{}, error) {
	m := map[string]interface{}{
		"ip_address": d.Address,
		"cidr":       d.Cidr,
		"gateway":    d.Gateway,
	}

	return m, nil
}

type FloatingIp struct {
	Ipv4 FloatingIpv4 `json:"ipv4" yaml:"ipv4" toml:"ipv4"`
}

func (d *FloatingIp) marshalMap() (map[string]interface{}, error) {
	var err error
	m := map[string]interface{}{}
	m["ipv4"], err = d.Ipv4.marshalMap()
	if err != nil {
		return nil, err
	}

	return m, nil
}

type FloatingIpv4 struct {
	Active    bool       `json:"active"  yaml:"active" toml:"active"`
	IPAddress model.IPv4 `json:"ip_address,omitempty"  yaml:"ip_address,omitempty" toml:"ip_address,omitempty"`
}

func (d *FloatingIpv4) marshalMap() (map[string]interface{}, error) {
	m := map[string]interface{}{
		"active":     d.Active,
		"ip_address": d.IPAddress,
	}

	return m, nil
}

type DNS struct {
	Nameservers []Nameserver `json:"nameservers"  yaml:"nameservers" toml:"nameservers"`
}

func (d *DNS) marshalMap() (map[string]interface{}, error) {
	return nil, nil
}

type Features struct {
	DhcpEnabled bool `json:"dhcp_enabled"  yaml:"dhcp_enabled" toml:"dhcp_enabled"`
}

func (d *Features) marshalMap() (map[string]interface{}, error) {
	return nil, nil
}

type Nameserver struct {
	Host string
	Port uint16
}

func (n *Nameserver) String() string {
	if n.Port == 0 || n.Port == 53 {
		return n.Host
	}

	return net.JoinHostPort(n.Host, strconv.FormatUint(uint64(n.Port), 10))
}

func (n *Nameserver) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}

func (n *Nameserver) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}

func (n Nameserver) MarshalYAML() (interface{}, error) {
	return n.String(), nil
}

func (n *Nameserver) UnmarshalJSON(data []byte) (err error) {
	var s string
	err = json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	if ip := net.ParseIP(s); ip != nil {
		n.Host = s
		n.Port = 53
		return nil
	}
	host, portStr, err := net.SplitHostPort(s)
	if err != nil {
		return err
	}
	if portStr != "" {
		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return err
		}
		n.Port = uint16(port)
	} else {
		n.Port = 53
	}

	n.Host = host

	return nil
}

func (n *Nameserver) UnmarshalYAML(value *yaml.Node) error {
	var s string
	err := value.Decode(&s)
	if err != nil {
		return err
	}
	if ip := net.ParseIP(s); ip != nil {
		n.Host = s
		n.Port = 53
		return nil
	}
	host, portStr, err := net.SplitHostPort(s)
	if err != nil {
		return err
	}
	if portStr != "" {
		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return err
		}
		n.Port = uint16(port)
	} else {
		n.Port = 53
	}

	n.Host = host

	return nil
}

func (n *Nameserver) UnmarshalTOML(data []byte) (err error) {
	var s string
	err = toml.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	if ip := net.ParseIP(s); ip != nil {
		n.Host = s
		n.Port = 53
		return nil
	}
	host, portStr, err := net.SplitHostPort(s)
	if err != nil {
		return err
	}
	if portStr != "" {
		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return err
		}
		n.Port = uint16(port)
	} else {
		n.Port = 53
	}

	n.Host = host

	return nil
}
