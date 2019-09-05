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
package cmd

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/amari/cloud-metadata-server/pkg/core"
	"github.com/amari/cloud-metadata-server/pkg/metadataserver"
	"github.com/amari/cloud-metadata-server/pkg/store"
	"go.uber.org/zap"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// initialize the server core
		c, err := core.NewDevelopmentServer()
		if err != nil {
			panic("")
		}

		// bind metadataListeners
		if len(metadataBindAddrSlice) == 0 {
			c.Log().Fatal("needs at least one bind address")
		}
		metadataListeners := make([]*net.TCPListener, 0, len(metadataBindAddrSlice))
		for _, addr := range metadataBindAddrSlice {
			hostStr, portStr, err := net.SplitHostPort(addr)
			if err != nil {
				c.Log().Fatal("invalid bind address", zap.String("bindAddress", addr), zap.NamedError("error", err))
			}
			ip := net.ParseIP(hostStr)
			if ip == nil {
				c.Log().Fatal("invalid ip address", zap.String("bindAddress", addr))
			}
			port, err := strconv.ParseUint(portStr, 10, 16)
			if err != nil {
				c.Log().Fatal("invalid port", zap.String("bindAddress", addr))
			}
			metadataListener, err := net.ListenTCP("tcp4", &net.TCPAddr{
				IP:   ip,
				Port: int(port),
				Zone: "",
			})
			if err != nil {
				c.Log().Fatal("failed to create tcp listener", zap.String("bindAddress", addr), zap.NamedError("error", err))
			}
			c.Log().Info("started metadata server", zap.String("address", metadataListener.Addr().String()))
			metadataListeners = append(metadataListeners, metadataListener)
		}

		// initialize the store
		var s store.Store
		switch metadataStore {
		case "dir":
			dirStore, err := store.NewDirStore(c, metadataStoreDirCacheSize)
			if err != nil {
				c.Log().Fatal("failed to create directory store", zap.NamedError("error", err))
			}
			for _, path := range metadataStoreDirSlice {
				err := dirStore.AddPath(path)
				if err != nil {
					c.Log().Fatal("failed to create directory store", zap.NamedError("error", err), zap.String("path", path))
				}
			}
			s = dirStore
		default:
			c.Log().Fatal("unknown metadata store", zap.String("metadataStore", metadataStore))
		}

		// initialize the metadata server
		metadataSrvRoot, err := metadataserver.NewHTTPServer(c, s, neighborTableRefreshInterval)
		if err != nil {
			c.Log().Fatal("failed to create metadata server", zap.NamedError("error", err))
		}

		metadataSrv := http.Server{
			Handler: metadataSrvRoot,
		}
		for _, metadataListener := range metadataListeners {
			go func(listener net.Listener) {
				err = metadataSrv.Serve(listener)
				if err != nil {
					c.Log().Fatal("failed to create metadata server", zap.NamedError("error", err))
				}
			}(metadataListener)
		}

		// initialize the api server

		// wait for shutdown
		waitForShutdown(&metadataSrv, nil)
	},
}

var metadataBindAddrSlice []string
var metadataStore string
var metadataStoreDirSlice []string
var metadataStorePostgres string
var metadataStoreDirCacheSize int
var neighborTableRefreshInterval time.Duration

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	serveCmd.Flags().StringSliceVar(&metadataBindAddrSlice, "metadata-bind-addr", []string{"169.254.169.254:80"}, "")
	serveCmd.Flags().StringVar(&metadataStore, "metadata-store", "", "dir, postgres, mariadb, mysql")
	serveCmd.Flags().StringSliceVar(&metadataStoreDirSlice, "metadata-store-dir", nil, "")
	serveCmd.Flags().StringVar(&metadataStorePostgres, "metadata-store-postgres", "", "")
	serveCmd.Flags().IntVar(&metadataStoreDirCacheSize, "metadata-store-dir-cache-size", 128, "")
	serveCmd.Flags().DurationVar(&neighborTableRefreshInterval, "neighbor-table-refresh-interval", 1*time.Millisecond, "")
}

func waitForShutdown(metadataSrv *http.Server, apiSrv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-interruptChan

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if apiSrv != nil {
		apiSrv.Shutdown(ctx)
	}
	if metadataSrv != nil {
		metadataSrv.Shutdown(ctx)
	}

	zap.L().Sync()

	os.Exit(0)
}
