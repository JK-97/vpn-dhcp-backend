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

	router = r.PathPrefix("/api/v1").Subrouter()
	appendDHCPRouter(router, &appconfig.DHCPConfig, appconfig.PrivateKey)

	if appconfig.MongoURI != "" && appconfig.PublicKey != nil {
		router = r.PathPrefix("/api/v1/bootstrap").Subrouter()
		appendBootstrapRouter(router, appconfig.MongoURI, appconfig.PublicKey)
	}

	http.ListenAndServe(appconfig.Addr, r)
}
