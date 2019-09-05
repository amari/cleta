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
package net

import (
	"encoding/json"
	"net"

	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

type IPv4 net.IP

func (addr IPv4) String() string {
	return net.IP(addr).String()
}

func (addr IPv4) MarshalText() ([]byte, error) {
	return []byte(addr.String()), nil
}

func (addr IPv4) MarshalJSON() ([]byte, error) {
	return json.Marshal(addr.String())
}

func (addr IPv4) MarshalYAML() (interface{}, error) {
	return addr.String(), nil
}

func (addr *IPv4) UnmarshalJSON(data []byte) (err error) {
	var s string
	err = json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	ip := net.ParseIP(s).To4()
	if ip == nil {
		panic("")
		return nil
	}

	*addr = IPv4(ip)

	return nil
}

func (addr *IPv4) UnmarshalYAML(value *yaml.Node) error {
	var s string
	err := value.Decode(&s)
	if err != nil {
		return err
	}

	ip := net.ParseIP(s).To4()
	if ip == nil {
		panic("")
		return nil
	}

	*addr = IPv4(ip)

	return nil
}

func (addr *IPv4) UnmarshalTOML(data []byte) (err error) {
	var s string
	err = toml.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	ip := net.ParseIP(s).To4()
	if ip == nil {
		panic("")
		return nil
	}

	*addr = IPv4(ip)

	return nil
}

type IPv4Mask net.IPMask

func (addr IPv4Mask) String() string {
	return net.IP(addr).String()
}

func (addr IPv4Mask) MarshalText() ([]byte, error) {
	return []byte(addr.String()), nil
}

func (addr IPv4Mask) MarshalJSON() ([]byte, error) {
	return json.Marshal(addr.String())
}

func (addr IPv4Mask) MarshalYAML() (interface{}, error) {
	return addr.String(), nil
}

func (addr *IPv4Mask) UnmarshalJSON(data []byte) (err error) {
	var s string
	err = json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	ip := net.ParseIP(s).To4()
	if ip == nil {
		panic("")
		return nil
	}

	*addr = IPv4Mask(ip)

	return nil
}

func (addr *IPv4Mask) UnmarshalYAML(value *yaml.Node) error {
	var s string
	err := value.Decode(&s)
	if err != nil {
		return err
	}

	ip := net.ParseIP(s).To4()
	if ip == nil {
		panic("")
		return nil
	}

	*addr = IPv4Mask(ip)

	return nil
}

func (addr *IPv4Mask) UnmarshalTOML(data []byte) (err error) {
	var s string
	err = toml.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	ip := net.ParseIP(s).To4()
	if ip == nil {
		panic("")
		return nil
	}

	*addr = IPv4Mask(ip)

	return nil
}
