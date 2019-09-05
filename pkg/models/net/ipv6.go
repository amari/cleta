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
	"strconv"

	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

type IPv6PrefixLen uint8

// UnmarshalJSON implements `json.Unmarshaler`
func (l *IPv6PrefixLen) UnmarshalJSON(data []byte) error {
	var a uint8
	err := json.Unmarshal(data, &a)
	if err != nil {
		var b string
		err = json.Unmarshal(data, &b)
		if err != nil {
			return err
		}
		v, err := strconv.ParseUint(b, 10, 8)
		if err != nil {
			return err
		}
		a = uint8(v)
	}
	*l = IPv6PrefixLen(a)

	return nil
}

// UnmarshalText implements `encoding.Unmarshaler`
func (l *IPv6PrefixLen) UnmarshalText(data []byte) error {
	v, err := strconv.ParseUint(string(data), 10, 8)
	if err != nil {
		return nil
	}
	*l = IPv6PrefixLen(uint8(v))

	return nil
}

// UnmarshalYAML implements `yaml.Unmarshaler`
func (l *IPv6PrefixLen) UnmarshalYAML(value *yaml.Node) error {
	var a uint8
	err := value.Decode(&a)
	if err != nil {
		var b string
		err = value.Decode(&b)
		if err != nil {
			return err
		}
		v, err := strconv.ParseUint(b, 10, 8)
		if err != nil {
			return err
		}
		a = uint8(v)
	}
	*l = IPv6PrefixLen(a)

	return nil
}

// UnmarshalTOML implements `toml.Unmarshaler`
func (l *IPv6PrefixLen) UnmarshalTOML(data []byte) error {
	var a uint8
	err := toml.Unmarshal(data, &a)
	if err != nil {
		var b string
		err = toml.Unmarshal(data, &b)
		if err != nil {
			return err
		}
		v, err := strconv.ParseUint(b, 10, 8)
		if err != nil {
			return err
		}
		a = uint8(v)
	}
	*l = IPv6PrefixLen(a)

	return nil
}

type IPv6 net.IP

func (addr IPv6) String() string {
	return net.IP(addr).String()
}

func (addr IPv6) MarshalText() ([]byte, error) {
	return []byte(addr.String()), nil
}

func (addr IPv6) MarshalJSON() ([]byte, error) {
	return json.Marshal(addr.String())
}

func (addr IPv6) MarshalYAML() (interface{}, error) {
	return addr.String(), nil
}

func (addr *IPv6) UnmarshalJSON(data []byte) (err error) {
	var s string
	err = json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	ip := net.ParseIP(s)
	if ip == nil {
		panic("")
		return nil
	}
	if x := ip.To4(); x != nil {
		panic("")
		return nil
	}

	*addr = IPv6(ip)

	return nil
}

func (addr *IPv6) UnmarshalYAML(value *yaml.Node) error {
	var s string
	err := value.Decode(&s)
	if err != nil {
		return err
	}

	ip := net.ParseIP(s)
	if ip == nil {
		panic("")
		return nil
	}
	if x := ip.To4(); x != nil {
		panic("")
		return nil
	}

	*addr = IPv6(ip)

	return nil
}

func (addr *IPv6) UnmarshalTOML(data []byte) (err error) {
	var s string
	err = toml.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	ip := net.ParseIP(s)
	if ip == nil {
		panic("")
		return nil
	}
	if x := ip.To4(); x != nil {
		panic("")
		return nil
	}

	*addr = IPv6(ip)

	return nil
}
