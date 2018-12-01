package storage

import (
	"context"

	"ploy.codes/microcloud/metadata/digitalocean/v1"
)

type Storage interface {
	GetMetadata(ctx context.Context, hwaddr string) (*Metadata, error)
}

type Metadata struct {
	Kind string `json:"kind" yaml:"kind"`

	DigitalOceanV1Droplet *v1.Droplet `json:"metadata" yaml:"metadata"`
}
