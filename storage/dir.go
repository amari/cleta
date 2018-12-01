package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	yaml "gopkg.in/yaml.v2"
	"ploy.codes/microcloud/metadata/digitalocean/v1"
)

type Dir struct {
	// 1. enumerate all files in directory
	// 2. filter files by extension or rule
	// 3. produce a DirEntry for each file
	// 4. start a goroutine to watch the directory for changes
	// 5. take lock and update the changed files
	Path string

	EntriesByPath   map[string]*DirEntry
	EntriesByHwAddr map[string]*DirEntry
	EntriesMutex    sync.Mutex
	PathWatcher     *fsnotify.Watcher
	DoneCh          chan<- bool
}

func ReadMetadataFile(path string) (*MetadataFile, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	metadataFile := &MetadataFile{}
	err = UnmarshalExt(filepath.Ext(path), buf, metadataFile)
	if err != nil {
		return nil, err
	}

	return metadataFile, nil
}

func OpenDir(path string) (*Dir, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	dir := &Dir{
		Path:            path,
		EntriesByPath:   make(map[string]*DirEntry, len(files)),
		EntriesByHwAddr: make(map[string]*DirEntry, len(files)),
	}

	for _, file := range files {
		dirEntry := &DirEntry{
			Path: filepath.Join(path, file.Name()),
		}

		metadataFile, err := ReadMetadataFile(dirEntry.Path)

		if err != nil {
			dirEntry.Err = err
			continue
		}

		dirEntry.Kind = metadataFile.Kind
		dirEntry.HardwareAddrs, err = metadataFile.HardwareAddrs()
		dirEntry.Metadata = &metadataFile.Metadata

		if err != nil {
			log.Println(err)
		}

		for _, hwaddr := range dirEntry.HardwareAddrs {
			hwaddrString := hwaddr.String()

			dir.EntriesByHwAddr[hwaddrString] = dirEntry
		}

		dir.EntriesByPath[dirEntry.Path] = dirEntry
	}

	// TODO: start a goroutine that watches for directory changes
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	doneCh := make(chan bool)

	dir.PathWatcher = watcher
	dir.DoneCh = doneCh

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("created file:", event.Name)
					path := filepath.Join(path, event.Name)
					// read the file
					metadataFile, err := ReadMetadataFile(path)
					if err != nil {
						log.Println(err)
						continue
					}

					dir.EntriesMutex.Lock()

					// insert the new entries
					dirEntry := &DirEntry{
						Path: path,
					}

					dirEntry.Kind = metadataFile.Kind
					dirEntry.HardwareAddrs, err = metadataFile.HardwareAddrs()
					dirEntry.Metadata = &metadataFile.Metadata

					if err != nil {
						log.Println(err)
						dirEntry.Err = err
						dir.EntriesMutex.Unlock()
						continue
					}

					for _, hwaddr := range dirEntry.HardwareAddrs {
						hwaddrString := hwaddr.String()

						dir.EntriesByHwAddr[hwaddrString] = dirEntry
					}

					dir.EntriesByPath[dirEntry.Path] = dirEntry
					dir.EntriesMutex.Unlock()
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)

					// reread the file
					path := filepath.Join(path, event.Name)
					metadataFile, err := ReadMetadataFile(path)
					if err != nil {
						log.Println(err)
						continue
					}

					// TODO: has anything changed?

					dir.EntriesMutex.Lock()

					// remove old entries
					if entry, ok := dir.EntriesByPath[filepath.Join(dir.Path, event.Name)]; ok {
						//
						for _, hwaddr := range entry.HardwareAddrs {
							delete(dir.EntriesByHwAddr, hwaddr.String())
						}

						delete(dir.EntriesByPath, entry.Path)
					}

					// insert the new entries
					dirEntry := &DirEntry{
						Path: filepath.Join(path, event.Name),
					}

					dirEntry.Kind = metadataFile.Kind
					dirEntry.HardwareAddrs, err = metadataFile.HardwareAddrs()
					dirEntry.Metadata = &metadataFile.Metadata

					if err != nil {
						log.Println(err)
						dirEntry.Err = err
						dir.EntriesMutex.Unlock()
						continue
					}

					for _, hwaddr := range dirEntry.HardwareAddrs {
						hwaddrString := hwaddr.String()

						dir.EntriesByHwAddr[hwaddrString] = dirEntry
					}

					dir.EntriesByPath[dirEntry.Path] = dirEntry

					dir.EntriesMutex.Unlock()
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					log.Println("removed file:", event.Name)

					dir.EntriesMutex.Lock()

					// remove old entries
					if entry, ok := dir.EntriesByPath[filepath.Join(dir.Path, event.Name)]; ok {
						//
						for _, hwaddr := range entry.HardwareAddrs {
							delete(dir.EntriesByHwAddr, hwaddr.String())
						}

						delete(dir.EntriesByPath, entry.Path)
					}

					dir.EntriesMutex.Unlock()
				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
					log.Println("renamed file:", event.Name)

					dir.EntriesMutex.Lock()
					defer dir.EntriesMutex.Unlock()

					// TODO: we need to use the mac addresses to identify the file, or use the droplet id as a key
				}
				// chmod
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	if err = watcher.Add(dir.Path); err != nil {
		return nil, err
	}

	return dir, nil
}

func (d *Dir) Close() error {
	// TODO: stop a goroutine that watches for directory changes
	if err := d.PathWatcher.Close(); err != nil {
		return err
	}

	return nil
}

func (d *Dir) GetMetadata(ctx context.Context, hwaddr string) (*Metadata, error) {
	d.EntriesMutex.Lock()
	defer d.EntriesMutex.Unlock()

	if entry, ok := d.EntriesByHwAddr[hwaddr]; ok {
		if entry.Err != nil {
			return nil, entry.Err
		}

		return entry.Metadata, nil
	}

	return nil, errors.New("not found")
}

type DirEntry struct {
	Path          string
	Kind          string
	HardwareAddrs []net.HardwareAddr
	Metadata      *Metadata
	Err           error
}

type MetadataFile struct {
	Metadata
}

func (f *MetadataFile) HardwareAddrs() ([]net.HardwareAddr, error) {
	switch f.Kind {
	case "digitalocean.com/v1":
		return f.Metadata.DigitalOceanV1Droplet.HardwareAddrs()
	default:
		// unknown "metadata" kind. bad file
		return nil, fmt.Errorf("unknown metadata format: %v", f.Kind)
	}
}

func (m *MetadataFile) UnmarshalJSON(b []byte) error {
	var tmp map[string]*json.RawMessage

	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}

	kindBytes, ok := tmp["kind"]
	if !ok {
		// key "kind" not found. bad file.
		return errors.New("bad file: kind not found")
	}

	var kind string
	if err := json.Unmarshal(*kindBytes, &kind); err != nil {
		return err
	}

	metadata, ok := tmp["metadata"]
	if !ok {
		// key "metadata" not found. bad file.
		return errors.New("bad file: metadata not found")
	}

	m.Metadata.Kind = kind

	switch kind {
	case "digitalocean.com/v1":
		droplet := v1.Droplet{}
		if err := json.Unmarshal(*metadata, &droplet); err != nil {
			return err
		}

		m.Metadata.DigitalOceanV1Droplet = &droplet
	default:
		// unknown "metadata" kind. bad file
		return fmt.Errorf("unknown metadata format: %v", kind)
	}

	return nil
}

func (m *MetadataFile) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var mapSlice yaml.MapSlice

	if err := unmarshal(&mapSlice); err != nil {
		return err
	}

	for _, item := range mapSlice {
		if item.Key == "kind" {
			switch item.Value {
			case "digitalocean.com/v1":
				metadata := struct {
					Kind string `json:"kind" yaml:"kind"`

					Metadata *v1.Droplet `json:"metadata" yaml:"metadata"`
				}{
					Metadata: &v1.Droplet{},
				}
				if err := unmarshal(&metadata); err != nil {
					log.Println(err)
					return err
				}

				m.Metadata.Kind = "digitalocean.com/v1"
				m.Metadata.DigitalOceanV1Droplet = metadata.Metadata
			default:
				return fmt.Errorf("unknown metadata format: %v", item.Key)
			}

			fmt.Printf("%+v", *m.DigitalOceanV1Droplet)
			return nil
		}
	}

	return errors.New("bad file: kind not found")
}

func UnmarshalExt(ext string, b []byte, v interface{}) error {
	if len(ext) == 0 {
		return fmt.Errorf("unknown codec: %v", ext)
	}

	switch ext[1:] {
	case "json":
		return json.Unmarshal(b, v)
	case "yaml", "yml":
		return yaml.Unmarshal(b, v)
	default:
		return fmt.Errorf("unknown codec: %v", ext[1:])
	}

	return nil
}
