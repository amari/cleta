package digitalocean

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	yaml "gopkg.in/yaml.v2"
	"ploy.codes/microcloud/storage"
)

type ApiV1 struct {
	*httprouter.Router

	storage.Storage
}

func NewApiV1(storage storage.Storage) ApiV1 {
	router := httprouter.New()

	router.HandlerFunc("GET", "/metadata/v1", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(`id
hostname
user-data
vendor-data
public-keys
region
interfaces/
dns/
floating_ip/
tags/
features/`))
	})

	router.Handle("GET", "/metadata/v1.:ext", func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		var (
			buf []byte
		)

		switch params[0].Value {
		case "json":
			buf, err = json.Marshal(m.DigitalOceanV1Droplet)
		case "yaml", "yml":
			buf, err = yaml.Marshal(m.DigitalOceanV1Droplet)
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(buf)
	})

	router.HandlerFunc("GET", "/metadata/v1/id", func(w http.ResponseWriter, req *http.Request) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.ID))
	})

	router.HandlerFunc("GET", "/metadata/v1/hostname", func(w http.ResponseWriter, req *http.Request) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.Hostname))
	})

	router.HandlerFunc("GET", "/metadata/v1/user-data", func(w http.ResponseWriter, req *http.Request) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.UserData))
	})

	router.HandlerFunc("GET", "/metadata/v1/vendor-data", func(w http.ResponseWriter, req *http.Request) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.VendorData))
	})

	router.HandlerFunc("GET", "/metadata/v1/public-keys", func(w http.ResponseWriter, req *http.Request) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		for _, publicKey := range m.DigitalOceanV1Droplet.PublicKeys {
			w.Write([]byte(publicKey))
			w.Write([]byte("\n"))
		}
	})

	router.HandlerFunc("GET", "/metadata/v1/region", func(w http.ResponseWriter, req *http.Request) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.Region))
	})

	router.HandlerFunc("GET", "/metadata/v1/interfaces/", func(w http.ResponseWriter, req *http.Request) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		if len(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces) > 0 {
			w.Write([]byte("public/\n"))
		}
		if len(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces) > 0 {
			w.Write([]byte("private/\n"))
		}
	})

	router.HandlerFunc("GET", "/metadata/v1/interfaces/public/", func(w http.ResponseWriter, req *http.Request) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		if len(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces) == 0 {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		for i := range m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces {
			w.Write([]byte(fmt.Sprintf("%v/", i)))
		}
	})

	router.HandlerFunc("GET", "/metadata/v1/interfaces/private/", func(w http.ResponseWriter, req *http.Request) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		if len(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces) == 0 {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		for i := range m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces {
			w.Write([]byte(fmt.Sprintf("%v/", i)))
		}
	})

	router.Handle("GET", "/metadata/v1/interfaces/public/:interface_id/", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces)) <= interfaceID {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(`mac
type
ipv4/
ipv6/
anchor_id/`))
	})
	router.Handle("GET", "/metadata/v1/interfaces/public/:interface_id/mac", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces)) <= interfaceID {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].Mac))
	})
	router.Handle("GET", "/metadata/v1/interfaces/public/:interface_id/type", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces)) <= interfaceID {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte("public"))
	})
	router.Handle("GET", "/metadata/v1/interfaces/public/:interface_id/ipv4/", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].Ipv4 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(`address
netmask
gateway`))
	})
	router.Handle("GET", "/metadata/v1/interfaces/public/:interface_id/ipv4/address", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].Ipv4 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].Ipv4.Address))
	})
	router.Handle("GET", "/metadata/v1/interfaces/public/:interface_id/ipv4/netmask", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].Ipv4 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].Ipv4.Netmask))
	})
	router.Handle("GET", "/metadata/v1/interfaces/public/:interface_id/ipv4/gateway", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].Ipv4 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].Ipv4.Gateway))
	})
	router.Handle("GET", "/metadata/v1/interfaces/public/:interface_id/ipv6/", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].Ipv6 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(`address
cidr
gateway`))
	})
	router.Handle("GET", "/metadata/v1/interfaces/public/:interface_id/ipv6/address", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].Ipv6 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].Ipv6.Address))
	})
	router.Handle("GET", "/metadata/v1/interfaces/public/:interface_id/ipv6/cidr", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].Ipv6 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].Ipv6.Cidr))
	})
	router.Handle("GET", "/metadata/v1/interfaces/public/:interface_id/ipv6/gateway", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].Ipv6 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].Ipv6.Gateway))
	})
	router.Handle("GET", "/metadata/v1/interfaces/public/:interface_id/anchor_ipv4/", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].AnchorIpv4 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(`address
netmask
gateway`))
	})
	router.Handle("GET", "/metadata/v1/interfaces/public/:interface_id/anchor_ipv4/address", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].AnchorIpv4 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].AnchorIpv4.Address))
	})
	router.Handle("GET", "/metadata/v1/interfaces/public/:interface_id/anchor_ipv4/netmask", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].AnchorIpv4 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].AnchorIpv4.Netmask))
	})
	router.Handle("GET", "/metadata/v1/interfaces/public/:interface_id/anchor_ipv4/gateway", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].AnchorIpv4 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.NetworkInterfaces.PublicInterfaces[interfaceID].AnchorIpv4.Gateway))
	})
	router.Handle("GET", "/metadata/v1/interfaces/private/:interface_id/", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces)) <= interfaceID {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(`mac
type
ipv4/
ipv6/`))
	})
	router.Handle("GET", "/metadata/v1/interfaces/private/:interface_id/mac", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces)) <= interfaceID {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces[interfaceID].Mac))
	})
	router.Handle("GET", "/metadata/v1/interfaces/private/:interface_id/type", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces)) <= interfaceID {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte("private"))
	})
	router.Handle("GET", "/metadata/v1/interfaces/private/:interface_id/ipv4/", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces[interfaceID].Ipv4 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(`address
netmask
gateway`))
	})
	router.Handle("GET", "/metadata/v1/interfaces/private/:interface_id/ipv4/address", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces[interfaceID].Ipv4 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces[interfaceID].Ipv4.Address))
	})
	router.Handle("GET", "/metadata/v1/interfaces/private/:interface_id/ipv4/netmask", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces[interfaceID].Ipv4 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces[interfaceID].Ipv4.Netmask))
	})
	router.Handle("GET", "/metadata/v1/interfaces/private/:interface_id/ipv4/gateway", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces[interfaceID].Ipv4 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces[interfaceID].Ipv4.Gateway))
	})
	router.Handle("GET", "/metadata/v1/interfaces/private/:interface_id/ipv6/", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces[interfaceID].Ipv6 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(`address
cidr
gateway`))
	})
	router.Handle("GET", "/metadata/v1/interfaces/private/:interface_id/ipv6/address", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces[interfaceID].Ipv6 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces[interfaceID].Ipv6.Address))
	})
	router.Handle("GET", "/metadata/v1/interfaces/private/:interface_id/ipv6/cidr", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces[interfaceID].Ipv6 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces[interfaceID].Ipv6.Cidr))
	})
	router.Handle("GET", "/metadata/v1/interfaces/private/:interface_id/ipv6/gateway", func(w http.ResponseWriter, req *http.Request, param httprouter.Params) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		interfaceID, err := strconv.ParseUint(param[0].Value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if uint64(len(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces)) <= interfaceID || m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces[interfaceID].Ipv6 == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.NetworkInterfaces.PrivateInterfaces[interfaceID].Ipv6.Gateway))
	})
	router.HandlerFunc("GET", "/metadata/v1/floating_ip/", func(w http.ResponseWriter, req *http.Request) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		if m.DigitalOceanV1Droplet.FloatingIP == nil {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		w.Write([]byte(`ipv4`))
	})
	router.HandlerFunc("GET", "/metadata/v1/floating_ip/ipv4/", func(w http.ResponseWriter, req *http.Request) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		if m.DigitalOceanV1Droplet.FloatingIP == nil {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		w.Write([]byte(`active
	ip_address`))
	})
	router.HandlerFunc("GET", "/metadata/v1/floating_ip/ipv4/active", func(w http.ResponseWriter, req *http.Request) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		if m.DigitalOceanV1Droplet.FloatingIP == nil {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		if m.DigitalOceanV1Droplet.FloatingIP.Ipv4.Active {
			w.Write([]byte("true"))
		} else {
			w.Write([]byte("false"))
		}
	})
	router.HandlerFunc("GET", "/metadata/v1/floating_ip/ipv4/ip_address", func(w http.ResponseWriter, req *http.Request) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		if m.DigitalOceanV1Droplet.FloatingIP == nil {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		w.Write([]byte(m.DigitalOceanV1Droplet.FloatingIP.Ipv4.IPAddress))
	})
	router.HandlerFunc("GET", "/metadata/v1/dns/", func(w http.ResponseWriter, req *http.Request) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		if m.DigitalOceanV1Droplet.DNS == nil {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		w.Write([]byte(`nameservers`))
	})
	router.HandlerFunc("GET", "/metadata/v1/dns/nameservers", func(w http.ResponseWriter, req *http.Request) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		for _, nameserver := range m.DigitalOceanV1Droplet.DNS.Nameservers {
			w.Write([]byte(nameserver))
		}
	})
	router.HandlerFunc("GET", "/metadata/v1/tags", func(w http.ResponseWriter, req *http.Request) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		for _, tag := range m.DigitalOceanV1Droplet.Tags {
			w.Write([]byte(tag))
		}
	})
	router.HandlerFunc("GET", "/metadata/v1/features/", func(w http.ResponseWriter, req *http.Request) {
		_, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		w.Write([]byte(`dhcp_enabled`))
	})
	router.HandlerFunc("GET", "/metadata/v1/features/dhcp_enabled", func(w http.ResponseWriter, req *http.Request) {
		m, err := storage.GetMetadata(context.Background(), req.RemoteAddr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		if m.DigitalOceanV1Droplet.Features.DhcpEnabled {
			w.Write([]byte("true"))
		} else {
			w.Write([]byte("false"))
		}
	})

	return ApiV1{
		Router:  router,
		Storage: storage,
	}
}
