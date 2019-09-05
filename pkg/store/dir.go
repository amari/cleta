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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/amari/cloud-metadata-server/pkg/core"
	"github.com/amari/cloud-metadata-server/pkg/models/document"
	"github.com/amari/cloud-metadata-server/pkg/models/metadata"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"

	lru "github.com/hashicorp/golang-lru"
)

// A DirStore is a store backed by one or more filesystem directories.
type DirStore struct {
	*core.Server

	doneCh chan struct{}

	m *sync.RWMutex

	watcher *fsnotify.Watcher
	paths   map[string]struct{}
	// CanonicalDataLinkAddr to []typeURIsForDataLinkAddr
	typeURIsForDataLinkAddr map[string][]string
	// FilePath to []CanonicalDataLinkAddr
	dataLinkAddrsForFilePath map[string][]string
	// (CanonicalDataLinkAddr, TypeURI) to FilePath
	filePathForDataLinkAddrAndTypeURI map[string]map[string]string
	// Cache FilePath to model.Metadata
	documentCache *lru.ARCCache

	// CanonicalDataLinkAddr to FilePath
	// filePathForDataLinkAddr map[string]string
}

func NewDirStore(c *core.Server, cacheSize int) (*DirStore, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	cache, err := lru.NewARC(cacheSize)
	if err != nil {
		return nil, err
	}

	store := &DirStore{
		Server: c,

		doneCh:                            make(chan struct{}),
		m:                                 &sync.RWMutex{},
		watcher:                           w,
		paths:                             map[string]struct{}{},
		typeURIsForDataLinkAddr:           map[string][]string{},
		dataLinkAddrsForFilePath:          map[string][]string{},
		filePathForDataLinkAddrAndTypeURI: map[string]map[string]string{},
		documentCache:                     cache,
	}

	go func(s *DirStore) {
		defer s.watcher.Close()
		for {
			select {
			case <-s.doneCh:
				return
			case event, _ := <-s.watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					s.didAddFile(event.Name)
				} else if event.Op&fsnotify.Chmod == fsnotify.Chmod || event.Op&fsnotify.Write == fsnotify.Write {
					s.didChangeFile(event.Name)
				} else if event.Op&fsnotify.Rename == fsnotify.Rename || event.Op&fsnotify.Remove == fsnotify.Remove {
					s.didRemoveFile(event.Name)
				}
			}
		}
	}(store)

	return store, nil
}

func (s *DirStore) Close(path string) error {
	close(s.doneCh)

	return nil
}

func (s *DirStore) AddPath(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return err
	}

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() {
			s.didAddFile(path)
		}
		return err
	})

	if err := s.watcher.Add(path); err != nil {
		return err
	}

	return nil
}

func (s *DirStore) didAddFile(path string) {
	file, err := readDocumentFromFile(path)
	if err != nil {
		//s.Log().Error(err.Error())
		return
	}
	canonicalDataLinkAddrs := []string{}

	s.m.Lock()
	defer s.m.Unlock()

	for _, dataLinkAddr := range file.Contents.DataLinkAddrs() {
		canonicalDataLinkAddr := dataLinkAddr.CanonicalString()
		canonicalDataLinkAddrs = append(canonicalDataLinkAddrs, canonicalDataLinkAddr)
		s.typeURIsForDataLinkAddr[canonicalDataLinkAddr] = append(s.typeURIsForDataLinkAddr[canonicalDataLinkAddr], file.TypeURI())
		if filePathForTypeURI, ok := s.filePathForDataLinkAddrAndTypeURI[canonicalDataLinkAddr]; ok {
			filePathForTypeURI[file.TypeURI()] = path
		} else {
			s.filePathForDataLinkAddrAndTypeURI[canonicalDataLinkAddr] = map[string]string{
				file.TypeURI(): path,
			}
		}
		s.documentCache.Remove(canonicalDataLinkAddr)
	}
	s.dataLinkAddrsForFilePath[path] = canonicalDataLinkAddrs

	fmt.Printf("didAddFile(%v)\n", path)
}

func (s *DirStore) didChangeFile(path string) {
	file, err := readDocumentFromFile(path)
	if err != nil {
		//s.Log().Error(err.Error())
		return
	}

	s.m.Lock()
	defer s.m.Unlock()

	s.documentCache.Remove(path)

	// remove
	if dataLinkAddrs, ok := s.dataLinkAddrsForFilePath[path]; ok {
		for _, dataLinkAddr := range dataLinkAddrs {
			delete(s.typeURIsForDataLinkAddr, dataLinkAddr)
			delete(s.filePathForDataLinkAddrAndTypeURI, dataLinkAddr)
		}
		delete(s.dataLinkAddrsForFilePath, path)
	}

	// add

	canonicalDataLinkAddrs := []string{}

	for _, dataLinkAddr := range file.Contents.DataLinkAddrs() {
		canonicalDataLinkAddr := dataLinkAddr.CanonicalString()
		canonicalDataLinkAddrs = append(canonicalDataLinkAddrs, canonicalDataLinkAddr)
		s.typeURIsForDataLinkAddr[canonicalDataLinkAddr] = append(s.typeURIsForDataLinkAddr[canonicalDataLinkAddr], file.TypeURI())
		if filePathForTypeURI, ok := s.filePathForDataLinkAddrAndTypeURI[canonicalDataLinkAddr]; ok {
			filePathForTypeURI[file.TypeURI()] = path
		} else {
			s.filePathForDataLinkAddrAndTypeURI[canonicalDataLinkAddr] = map[string]string{
				file.TypeURI(): path,
			}
		}
	}
	s.dataLinkAddrsForFilePath[path] = canonicalDataLinkAddrs

	fmt.Printf("didChangeFile(%v)\n", path)
}

func (s *DirStore) didRemoveFile(path string) {
	s.m.Lock()
	defer s.m.Unlock()

	// map path to dataLinkAddrs
	if dataLinkAddrs, ok := s.dataLinkAddrsForFilePath[path]; ok {
		for _, dataLinkAddr := range dataLinkAddrs {
			delete(s.typeURIsForDataLinkAddr, dataLinkAddr)
			delete(s.filePathForDataLinkAddrAndTypeURI, dataLinkAddr)
		}
		delete(s.dataLinkAddrsForFilePath, path)
	}
	s.documentCache.Remove(path)

	fmt.Printf("didRemoveFile(%v)\n", path)
}

func (s *DirStore) getDocument(path string) (*document.Document, error) {
	if v, ok := s.documentCache.Get(path); ok {
		if m, ok := v.(*document.Document); ok {
			return m, nil
		}
	}
	// write-back update the cache
	file, err := readDocumentFromFile(path)
	if err != nil {
		return nil, err
	}
	s.documentCache.Add(path, file.Contents)

	return file, nil
}

func (s *DirStore) ListSupportedTypeURIs(ctx context.Context, canonicalDataLinkAddr string) ([]string, error) {
	if v, ok := s.typeURIsForDataLinkAddr[canonicalDataLinkAddr]; ok {
		return v, nil
	}

	return nil, ErrNotFound
}

func (s *DirStore) ListDocuments(ctx context.Context, canonicalDataLinkAddr string) ([]document.Document, error) {
	typeURIs, err := s.ListSupportedTypeURIs(ctx, canonicalDataLinkAddr)
	if err != nil {
		return nil, err
	}

	if filePathForTypeURI, ok := s.filePathForDataLinkAddrAndTypeURI[canonicalDataLinkAddr]; ok {
		filePaths := make(map[string]struct{}, len(typeURIs))
		ret := make([]document.Document, 0, len(typeURIs))

		for _, typeURI := range typeURIs {
			if filePath, ok := filePathForTypeURI[typeURI]; ok {
				filePaths[filePath] = struct{}{}
			}
		}
		// ensure that the file is unique
		for filePath := range filePaths {
			// get metadata from the cache
			d, err := s.getDocument(filePath)
			if err != nil {
				//return nil, err
				continue
			}
			ret = append(ret, *d)
		}

		return ret, nil
	}

	return nil, ErrNotFound
}

func (s *DirStore) GetDocument(ctx context.Context, canonicalDataLinkAddr string, typeURI string) (*document.Document, error) {
	if filePathForTypeURI, ok := s.filePathForDataLinkAddrAndTypeURI[canonicalDataLinkAddr]; ok {
		if filePath, ok := filePathForTypeURI[typeURI]; ok {
			// get metadata from the cache
			d, err := s.getDocument(filePath)
			if err != nil {
				return nil, err
			}
			return d, nil
		}
	}

	return nil, ErrNotFound
}

type rawMetadataFileJSON struct {
	TypeURI     string          `json:"kind"`
	RawMetadata json.RawMessage `json:"metadata"`
}

type rawMetadataFileYAML struct {
	TypeURI     string    `yaml:"kind"`
	RawMetadata yaml.Node `yaml:"metadata"`
}

type metadataFile struct {
	TypeURI  string
	Metadata document.Metadata
}

var errBadFileExtension = errors.New("Bad file extension")
var errBadPath = errors.New("Bad path")

func readDocumentFromFile(path string) (d *document.Document, err error) {
	// TODO: limit the file size
	data, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		err = errBadPath
		return
	}

	switch filepath.Ext(path) {
	case ".json":
		err = json.Unmarshal(data, &d)
		if err != nil {
			d = nil
			return
		}

		return
	case ".yaml":
		fallthrough
	case ".yml":
		err = yaml.Unmarshal(data, &d)
		if err != nil {
			d = nil
			return
		}

		return
	default:
		err = errBadFileExtension
		return
	}
}

func readAndParseMetadataFile(path string) (*metadataFile, error) {
	// TODO: limit the file size
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, ErrNotFound
	}

	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		var rawFile rawMetadataFileJSON
		if err := json.Unmarshal(data, &rawFile); err != nil {
			return nil, err
		}

		metadata, err := metadata.UnmarshalMetadataJSON(rawFile.TypeURI, rawFile.RawMetadata)
		if err != nil {
			return nil, err
		}

		return &metadataFile{
			TypeURI:  rawFile.TypeURI,
			Metadata: metadata,
		}, nil
	case ".yaml":
		fallthrough
	case ".yml":
		var rawFile rawMetadataFileYAML
		if err := yaml.Unmarshal(data, &rawFile); err != nil {
			return nil, err
		}

		metadata, err := metadata.UnmarshalMetadataYAML(rawFile.TypeURI, &rawFile.RawMetadata)
		if err != nil {
			return nil, err
		}

		return &metadataFile{
			TypeURI:  rawFile.TypeURI,
			Metadata: metadata,
		}, nil
	default:
		return nil, ErrNotFound
	}
}
