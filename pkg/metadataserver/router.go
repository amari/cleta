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

package metadataserver

import (
	"github.com/amari/cloud-metadata-server/pkg/core"
	"github.com/amari/cloud-metadata-server/pkg/metadataserver/digitalocean"
	"github.com/amari/cloud-metadata-server/pkg/store"
	"go.uber.org/zap"
)

type Router struct {
	*core.Server

	endpoints map[string]Endpoint
}

func NewRouter(c *core.Server, s store.Store) *Router {
	return &Router{
		Server: c,
		endpoints: map[string]Endpoint{
			digitalocean.TypeURIV1: digitalocean.NewEndpointV1(c.WithLoggerFields(zap.String("kind", digitalocean.TypeURIV1)), s),
		},
	}
}

func (r *Router) Match(typeURI string) Endpoint {
	if endpoint, ok := r.endpoints[typeURI]; ok {
		return endpoint
	}
	return nil
}
