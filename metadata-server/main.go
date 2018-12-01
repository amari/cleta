package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"ploy.codes/microcloud/apiserver/digitalocean"
	"ploy.codes/microcloud/storage"
)

func main() {
	storage, err := storage.OpenDir(".")
	if err != nil {
		panic(err)
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", ":http")
	if err != nil {
		panic(err)
	}

	lis, err := ListenTCP("tcp", tcpAddr)
	if err != nil {
		panic(err)
	}

	srv := http.Server{
		Handler: &Router{
			Storage:        storage,
			DigitalOceanV1: digitalocean.NewApiV1(storage),
		},
	}

	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}

type Router struct {
	Storage storage.Storage

	DigitalOceanV1 digitalocean.ApiV1
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m, err := r.Storage.GetMetadata(context.Background(), req.RemoteAddr)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)

		return
	}

	switch m.Kind {
	case "digitalocean.com/v1":
		r.DigitalOceanV1.ServeHTTP(w, req)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// IPv4 only
// todo, route to http handle based on the mac address
