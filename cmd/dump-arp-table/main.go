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

package main

import (
	"context"
	"fmt"

	"github.com/amari/cloud-metadata-server/internal/pkg/arp"
)

func main() {
	arpTable, _ := arp.NewTable()
	defer arpTable.Close()

	if err := arpTable.Poll(context.Background(), func(_ context.Context, entry arp.Entry) {
		fmt.Printf("%v: %v -> %v\n", entry.InterfaceName(), entry.RemoteIP(), entry.HardwareAddr())
	}); err != nil {
		panic(err)
	}
}
