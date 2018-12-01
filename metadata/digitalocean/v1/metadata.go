package v1

import (
	"net"
)

const TypeUri = "digitalocean.com/v1"

type Droplet struct {
	ID                string            `json:"droplet_id" yaml:"droplet_id"`
	Hostname          string            `json:"hostname" yaml:"hostname"`
	UserData          UserData          `json:"user_data"  yaml:"user_data"`
	VendorData        VendorData        `json:"vendor_data"  yaml:"vendor_data"`
	PublicKeys        []PublicKey       `json:"public_keys"  yaml:"public_keys"`
	Region            string            `json:"region"  yaml:"region"`
	NetworkInterfaces NetworkInterfaces `json:"interfaces" yaml:"interfaces"`
	FloatingIP        *FloatingIp       `json:"floating_ip"  yaml:"floating_ip"`
	DNS               *DNS              `json:"dns"  yaml:"dns"`
	Tags              []string          `json:"tags"  yaml:"tags"`
	Features          Features          `json:"features"  yaml:"features"`
}

var availableDropletKeys = [...]string{"id", "hostname", "user-data", "vendor-data", "public-keys", "region", "interfaces/", "dns/", "floating_ip/", "tags/", "features/"}

func (d *Droplet) availableKeys() []string {
	return availableDropletKeys[:]
}

func (d *Droplet) HardwareAddrs() ([]net.HardwareAddr, error) {
	res := make([]net.HardwareAddr, 0, len(d.NetworkInterfaces.PrivateInterfaces)+len(d.NetworkInterfaces.PublicInterfaces))

	for i := 0; i < len(d.NetworkInterfaces.PrivateInterfaces); i++ {
		addr, err := net.ParseMAC(d.NetworkInterfaces.PrivateInterfaces[i].Mac)
		if err != nil {
			return nil, err
		}

		res = append(res, addr)
	}

	for i := 0; i < len(d.NetworkInterfaces.PublicInterfaces); i++ {
		addr, err := net.ParseMAC(d.NetworkInterfaces.PublicInterfaces[i].Mac)
		if err != nil {
			return nil, err
		}

		res = append(res, addr)
	}

	return res, nil
}

type UserData string
type VendorData string
type PublicKey string

type NetworkInterfaces struct {
	PrivateInterfaces []PrivateNetworkInterface `json:"private" yaml:"private"`
	PublicInterfaces  []PublicNetworkInterface  `json:"public" yaml:"public"`
}

type PublicNetworkInterface struct {
	Mac        string    `json:"mac" yaml:"mac"`
	Ipv4       *Ipv4Addr `json:"ipv4" yaml:"ipv4"`
	Ipv6       *Ipv6Addr `json:"ipv6" yaml:"ipv6"`
	AnchorIpv4 *Ipv4Addr `json:"anchor_ipv4"  yaml:"anchor_ipv4"`
}

type PrivateNetworkInterface struct {
	Mac  string    `json:"mac" yaml:"mac"`
	Ipv4 *Ipv4Addr `json:"ipv4" yaml:"ipv4"`
	Ipv6 *Ipv6Addr `json:"ipv6" yaml:"ipv6"`
}

type Ipv4Addr struct {
	Address string `json:"address" yaml:"address"`
	Netmask string `json:"netmask" yaml:"netmask"`
	Gateway string `json:"gateway" yaml:"gateway"`
}

type Ipv6Addr struct {
	Address string `json:"address" yaml:"address"`
	Cidr    string `json:"cidr" yaml:"cidr"`
	Gateway string `json:"gateway" yaml:"gateway"`
}

type FloatingIp struct {
	Ipv4 FloatingIpv4 `json:"ipv4" yaml:"ipv4"`
}

type FloatingIpv4 struct {
	Active    bool   `json:"active"  yaml:"active"`
	IPAddress string `json:"ip_address"  yaml:"ip_address"`
}

type DNS struct {
	Nameservers []string `json:"nameservers"  yaml:"nameservers"`
}

type Features struct {
	DhcpEnabled bool `json:"dhcp_enabled"  yaml:"dhcp_enabled"`
}
