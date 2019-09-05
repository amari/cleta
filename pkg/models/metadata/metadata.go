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

package metadata

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"

	digitalocean_v1 "github.com/amari/cloud-metadata-server/pkg/models/digitalocean/v1"
	model "github.com/amari/cloud-metadata-server/pkg/models/net"
	"github.com/pelletier/go-toml"
)

type Metadata interface {
	DataLinkAddrs() []model.DataLinkAddr
	TypeURI() string
}

func UnmarshalMetadataJSON(typeURI string, data []byte) (metadata Metadata, err error) {
	switch typeURI {
	case digitalocean_v1.TypeURI:
		metadata = new(digitalocean_v1.Droplet)
	default:
		return nil, fmt.Errorf("unknown typeURI %v", typeURI)
	}
	err = json.Unmarshal(data, &metadata)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func UnmarshalMetadataYAML(typeURI string, value *yaml.Node) (metadata Metadata, err error) {
	switch typeURI {
	case digitalocean_v1.TypeURI:
		var m digitalocean_v1.Droplet
		err = value.Decode(&m)
		if err != nil {
			return nil, err
		}
		metadata = &m
	default:
		return nil, fmt.Errorf("unknown typeURI %v", typeURI)
	}
	//err = metadata.(yaml.Unmarshaler).UnmarshalYAML(&rawFile.RawMetadata)
	/*err = rawFile.RawMetadata.Decode(&metadata)
	if err != nil {
		return nil, err
	}*/
	return metadata, nil
}

func UnmarshalMetadataTOML(typeURI string, data []byte) (metadata Metadata, err error) {
	switch typeURI {
	case digitalocean_v1.TypeURI:
		metadata = new(digitalocean_v1.Droplet)
	default:
		return nil, fmt.Errorf("unknown typeURI %v", typeURI)
	}
	err = toml.Unmarshal(data, &metadata)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}
