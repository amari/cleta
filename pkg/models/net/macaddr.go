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
	"encoding/base64"
	"encoding/json"
	"net"

	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// MACAddr is a "media access control" address
type MACAddr net.HardwareAddr

// ParseMAC wraps `net.ParseMAC`
func ParseMAC(s string) (MACAddr, error) {
	addr, err := net.ParseMAC(s)
	if err != nil {
		return nil, err
	}

	return MACAddr(addr), nil
}

// CanonicalString implements `DataLinkAddr`.
func (m MACAddr) CanonicalString() string {
	return base64.StdEncoding.EncodeToString(m)
}

// Bytes implements `DataLinkAddr`.
func (m MACAddr) Bytes() []byte {
	return []byte(m)
}

// HumanReadableString implements `DataLinkAddr`.
func (m MACAddr) HumanReadableString() string {
	return (net.HardwareAddr)(m).String()
}

// MarshalJSON implements `json.Marshaler`
func (m MACAddr) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.HumanReadableString())
}

// MarshalText implements `encoding.Marshaler`
func (m MACAddr) MarshalText() ([]byte, error) {
	return []byte(m.HumanReadableString()), nil
}

// MarshalYAML implements `yaml.Marshaler`
func (m MACAddr) MarshalYAML() (interface{}, error) {
	return m.HumanReadableString(), nil
}

// MarshalTOML implements `toml.Marshaler`
func (m MACAddr) MarshalTOML() ([]byte, error) {
	return []byte(m.HumanReadableString()), nil
}

// UnmarshalJSON implements `json.Unmarshaler`
func (m *MACAddr) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	hardwareAddr, err := net.ParseMAC(s)
	if err != nil {
		return err
	}
	//copy(*m, hardwareAddr)
	*m = MACAddr(hardwareAddr)

	return nil
}

// UnmarshalText implements `encoding.Unmarshaler`
func (m *MACAddr) UnmarshalText(data []byte) error {
	hardwareAddr, err := net.ParseMAC(string(data))
	if err != nil {
		return err
	}
	*m = MACAddr(hardwareAddr)

	return nil
}

// UnmarshalYAML implements `yaml.Unmarshaler`
func (m *MACAddr) UnmarshalYAML(value *yaml.Node) error {
	var s string
	err := value.Decode(&s)
	if err != nil {
		return err
	}
	hardwareAddr, err := net.ParseMAC(s)
	if err != nil {
		return err
	}
	//copy(*m, hardwareAddr)
	*m = MACAddr(hardwareAddr)

	return nil
}

// UnmarshalTOML implements `toml.Unmarshaler`
func (m *MACAddr) UnmarshalTOML(data []byte) error {
	var s string
	err := toml.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	hardwareAddr, err := net.ParseMAC(s)
	if err != nil {
		return err
	}
	//copy(*m, hardwareAddr)
	*m = MACAddr(hardwareAddr)

	return nil
}
