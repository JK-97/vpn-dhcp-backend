package main

//go:generate go run version_generate.go
import (
	"net/http"

	"dhcp-backend/option"
	"dhcp-backend/serve"

	"github.com/gorilla/mux"
)

const (
	codingTable = "ABCDEFGHIJKLMNOabcdefghijklmnopqrstuvwxyzPQRSTUVWXYZ0123456789-_"
)

var appconfig *option.ApplicationConfig

func main() {
	appconfig = option.ReadConfigFile("dhcp-backend.cfg")

	r := mux.NewRouter()

	r.NotFoundHandler = serve.NewNotFoundHandler()
	r.Use(serve.SimpleLoggingMw)

	router := r.PathPrefix("/api/v1/dns").Subrouter()

	appendDNSRoute(router, &appconfig.DNSConfig)

	// router.PathPrefix("").Methods(http.MethodPost).HandlerFunc(backend.AddRecord)
	// router.PathPrefix("").Methods(http.MethodDelete).HandlerFunc(backend.RemoveRecord)
	// router.PathPrefix("").Methods(http.MethodGet).HandlerFunc(backend.FindRecord)

	// kvc, err := clientv3.New(clientv3.Config{
	// 	Endpoints: []string{
	// 		"10.55.2.114:2379",
	// 	}})

	// if err != nil {
	// 	panic(err)
	// }

	// agent := etcd.DNSAgent{
	// 	Client: kvc,
	// 	Prefix: "/skydns",
	// }

	// vb := serve.VirtualNetworkBackend{
	// 	GatewayClient: &etcd.VPNGatewayClient{
	// 		Prefix: "/gw",
	// 		Client: kvc,
	// 	},
	// 	WorkerClient: &etcd.WorkerClient{
	// 		Prefix: "/wk",
	// 		Client: kvc,
	// 	},
	// 	HTTPClient: http.DefaultClient,
	// 	AgentPort:  9095,
	// 	Agent:      &agent,
	// 	Encoding:   base64.NewEncoding(codingTable).WithPadding(base64.NoPadding),
	// }

	router = r.PathPrefix("/api/v1").Subrouter()
	appendDHCPRouter(router, &appconfig.DHCPConfig, appconfig.PrivateKey)

	if appconfig.MongoURI != "" && appconfig.PublicKey != nil {
		router = r.PathPrefix("/api/v1/bootstrap").Subrouter()
		appendBootstrapRouter(router, appconfig.MongoURI, appconfig.PublicKey)
	}

	http.ListenAndServe(":1054", r)
}
