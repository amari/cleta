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

package document

import (
	"encoding/json"
	"errors"

	digitaloceanv1 "github.com/amari/cloud-metadata-server/pkg/models/digitalocean/v1"
	"github.com/amari/cloud-metadata-server/pkg/models/net"
	"gopkg.in/yaml.v3"
)

type Document struct {
	Kind     string   `json:"kind" yaml:"kind" toml:"kind"`
	Contents Metadata `json:"metadata" yaml:"metadata" toml:"metadata"`
}

func (d *Document) TypeURI() string {
	return d.Kind
}

func (d *Document) UnmarshalJSON(data []byte) (err error) {
	var rawDocument rawJSONDocument
	if err := json.Unmarshal(data, &rawDocument); err != nil {
		return err
	}

	m, err := unmarshalJSONMetadata(rawDocument.Kind, rawDocument.Contents)
	if err != nil {
		return err
	}

	d.Kind = rawDocument.Kind
	d.Contents = m

	return nil
}

func (d *Document) UnmarshalYAML(node *yaml.Node) (err error) {
	var rawDocument rawYAMLDocument
	if err := node.Decode(&rawDocument); err != nil {
		return err
	}

	m, err := unmarshalYAMLMetadata(rawDocument.Kind, &rawDocument.Contents)
	if err != nil {
		return err
	}

	d.Kind = rawDocument.Kind
	d.Contents = m

	return nil
}

type rawJSONDocument struct {
	Kind     string          `json:"kind" yaml:"kind" toml:"kind"`
	Contents json.RawMessage `json:"metadata" yaml:"metadata" toml:"metadata"`
}

type rawYAMLDocument struct {
	Kind     string    `json:"kind" yaml:"kind" toml:"kind"`
	Contents yaml.Node `json:"metadata" yaml:"metadata" toml:"metadata"`
}

type Metadata interface {
	DataLinkAddrs() []net.DataLinkAddr
}

var errBadTypeURI = errors.New("Bad TypeURI")

func unmarshalJSONMetadata(kind string, data json.RawMessage) (metadata Metadata, err error) {
	switch kind {
	case digitaloceanv1.TypeURI:
		metadata = new(digitaloceanv1.Droplet)
	default:
		err = errBadTypeURI
		return
	}
	err = json.Unmarshal(data, &metadata)
	if err != nil {
		metadata = nil
	}
	return
}

func unmarshalYAMLMetadata(kind string, node *yaml.Node) (metadata Metadata, err error) {
	switch kind {
	case digitaloceanv1.TypeURI:
		var droplet digitaloceanv1.Droplet
		err = node.Decode(&droplet)
		if err != nil {
			return
		}
		metadata = &droplet
	default:
		err = errBadTypeURI
		return
	}
	return
}
