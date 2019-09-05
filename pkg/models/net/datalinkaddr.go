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
	"errors"
)

// DataLinkAddr represents a OSI Layer 2 data link address.
type DataLinkAddr interface {
	// Bytes returns the data link address as a byte sequence.
	Bytes() []byte

	// CanonicalString returns a canonicalized string representation.
	CanonicalString() string

	HumanReadableString() string
}

// ParseCanonicalAddr parses a string previously created with `CanonicalString`.
func ParseCanonicalAddr(s string) (DataLinkAddr, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	switch len(data) {
	case 8:
		// EUI-64
		fallthrough
	case 6:
		// IEEE 802 MAC-48
		// EUI-48
		return MACAddr(data), nil
	case 20:
		// 20-octet IP over InfiniBand link-layer address
		return MACAddr(data), nil
	default:
		return nil, errBadDataLinkAddr
	}
}

var errBadDataLinkAddr = errors.New("Bad data-link addr")
