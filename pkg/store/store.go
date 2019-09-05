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

package store

import (
	"context"
	"errors"

	"github.com/amari/cloud-metadata-server/pkg/models/document"
)

type Store interface {
	ListSupportedTypeURIs(ctx context.Context, dataLinkAddr string) ([]string, error)

	ListDocuments(ctx context.Context, dataLinkAddr string) ([]document.Document, error)
	GetDocument(ctx context.Context, dataLinkAddr string, typeURI string) (*document.Document, error)
	//CreateDocument(ctx context.Context, dataLinkAddr string, document document.Document) error

	//UpdateDocument(ctx context.Context, dataLinkAddr string, document document.Document) error

	//DeleteManyDocuments(ctx context.Context, dataLinkAddr string) error
	//DeleteDocument(ctx context.Context, dataLinkAddr string, typeURI string) error
}

var ErrNotFound = errors.New("Not found")

/*type Store interface {
	//TypeURIForHardwareAddr(ctx context.Context, addr net.HardwareAddr) (string, error)

	//MetadataForHardwareAddr(ctx context.Context, addr net.HardwareAddr) (string, document.Metadata, error)

	ListTypeURIs(ctx context.Context, canonicalDataLinkAddr string) ([]string, error)
	ListMetadata(ctx context.Context, canonicalDataLinkAddr string) ([]document.Metadata, error)
	GetMetadata(ctx context.Context, canonicalDataLinkAddr string, typeURI string) (document.Metadata, error)
}

// ListTypeURIs(ctx context.Context, base64EncodedDataLinkAddr string) ([]string, error)
// ListMetadata(ctx context.Context, base64EncodedDataLinkAddr string) ([]document.Metadata, error)
// GetMetadata(ctx context.Context, base64EncodedDataLinkAddr string, typeURI string) (document.Metadata, error)

// ListTypeURIs(ctx context.Context, canonicalDataLinkAddrStr string) ([]string, error)
// ListMetadata(ctx context.Context, canonicalDataLinkAddrStr string) ([]document.Metadata, error)
// GetMetadata(ctx context.Context, canonicalDataLinkAddrStr string, typeURI string) (document.Metadata, error)

var ErrNotFound = errors.New("not found")

// SliceStore combines the contents of multiple stores together
type SliceStore []Store

func (s SliceStore) ListTypeURIs(ctx context.Context, canonicalDataLinkAddr string) ([]string, error) {
	if len(s) == 0 {
		return nil, ErrNotFound
	}

	for _, store := range s {
		typeURIs, err := store.ListTypeURIs(ctx, canonicalDataLinkAddr)
		if err != nil {
			log.Println(err)
			continue
		}

		return typeURIs, nil
	}

	return nil, ErrNotFound
}

func (s SliceStore) ListMetadata(ctx context.Context, canonicalDataLinkAddr string) ([]document.Metadata, error) {
	if len(s) == 0 {
		return nil, ErrNotFound
	}

	for _, store := range s {
		metadatas, err := store.ListMetadata(ctx, canonicalDataLinkAddr)
		if err != nil {
			log.Println(err)
			continue
		}

		return metadatas, nil
	}

	return nil, ErrNotFound
}

func (s SliceStore) GetMetadata(ctx context.Context, canonicalDataLinkAddr string, typeURI string) (document.Metadata, error) {
	if len(s) == 0 {
		return nil, ErrNotFound
	}

	for _, store := range s {
		metadata, err := store.GetMetadata(ctx, canonicalDataLinkAddr, typeURI)
		if err != nil {
			log.Println(err)
			continue
		}

		return metadata, nil
	}

	return nil, ErrNotFound
}
*/
