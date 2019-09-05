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
package arp

import (
	"context"
	"net"
	"sync"
	"time"
)

type Watcher struct {
	hardwareAddrsByIP4 map[string]net.HardwareAddr

	table  *Table
	lock   *sync.RWMutex
	ticker *time.Ticker
	done   chan struct{}
}

func NewWatcher(d time.Duration) (*Watcher, error) {
	//timer := time.NewTimer(period)
	table, err := NewTable()
	if err != nil {
		return nil, err
	}

	hardwareAddrsByIP4 := map[string]net.HardwareAddr{}

	table.Poll(context.Background(), func(_ context.Context, entry Entry) {
		hardwareAddrsByIP4[entry.RemoteIP().String()] = entry.HardwareAddr()
	})

	watcher := &Watcher{
		hardwareAddrsByIP4: hardwareAddrsByIP4,
		table:              table,
		lock:               &sync.RWMutex{},
		ticker:             time.NewTicker(d),
		done:               make(chan struct{}, 1),
	}

	go func(t *Table, w *Watcher) {
		for {
			select {
			case <-watcher.ticker.C:
				watcher.ForcePoll()
			case <-watcher.done:
				return
			}
		}
	}(table, watcher)

	return watcher, nil
}

func (w *Watcher) Close() error {
	w.ticker.Stop()
	close(w.done)

	return nil
}

// used when there's a cache miss and we want to be sure
func (w *Watcher) ForcePoll() error {
	hardwareAddrsByIP4 := map[string]net.HardwareAddr{}

	err := w.table.Poll(context.Background(), func(_ context.Context, entry Entry) {
		hardwareAddrsByIP4[entry.RemoteIP().String()] = entry.HardwareAddr()
	})
	if err != nil {
		return err
	}

	w.lock.Lock()
	w.hardwareAddrsByIP4 = hardwareAddrsByIP4
	w.lock.Unlock()

	return nil
}

func (w *Watcher) HardwareAddrsByIP() map[string]net.HardwareAddr {
	w.lock.RLock()
	defer w.lock.RUnlock()

	return w.hardwareAddrsByIP4
}

func (w *Watcher) GetHardwareAddrForIP4(ip net.IP) net.HardwareAddr {
	key := ip.String()

	w.lock.RLock()
	defer w.lock.RUnlock()

	if addr, ok := w.hardwareAddrsByIP4[key]; ok {
		return addr
	}

	return nil
}
