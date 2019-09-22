package main

import (
	"crypto/rsa"
	"dhcp-backend/etcd"
	"dhcp-backend/go-utils/logger"
	"dhcp-backend/option"
	"dhcp-backend/serve"
	"encoding/base64"
	"net/http"

	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

func getEtcdClient(config *option.DNSConfig) (kvc *clientv3.Client, err error) {
	kvc, err = clientv3.New(clientv3.Config{
		Endpoints: config.Endpoints,
		Username:  config.Username,
		Password:  config.Password,
	})
	return
}

func appendDNSRoute(router *mux.Router, config *option.DNSConfig) *serve.DNSBackend {

	kvc, err := getEtcdClient(config)

	if err != nil {
		panic(err)
	}

	agent := etcd.DNSAgent{
		Client: kvc,
		Prefix: config.DNSPrefix,
	}

	backend := serve.DNSBackend{Agent: &agent}

	router.PathPrefix("").Methods(http.MethodPost).HandlerFunc(backend.AddRecord)
	router.PathPrefix("").Methods(http.MethodDelete).HandlerFunc(backend.RemoveRecord)
	router.PathPrefix("").Methods(http.MethodGet).HandlerFunc(backend.FindRecord)
	return &backend
}

func appendDHCPRouter(router *mux.Router, config *option.DHCPConfig, pri *rsa.PrivateKey) *serve.VirtualNetworkBackend {

	kvc, err := getEtcdClient(&config.DNSConfig)

	if err != nil {
		panic(err)
	}

	agent := etcd.DNSAgent{
		Client: kvc,
		Prefix: config.DNSPrefix,
	}

	vb := serve.VirtualNetworkBackend{
		PrivateKey: pri,
		GatewayClient: &etcd.VPNGatewayClient{
			Prefix: config.GatewayPrefix,
			Client: kvc,
		},
		WorkerClient: &etcd.WorkerClient{
			Prefix: config.WorkerPrefix,
			Client: kvc,
		},
		HTTPClient:   http.DefaultClient,
		AgentPort:    config.VPNAgentPort,
		Agent:        &agent,
		Encoding:     base64.NewEncoding(codingTable).WithPadding(base64.NoPadding),
		RegisterPath: config.RegisterPath,
	}

	router.PathPrefix("/openvpn/register").Methods(http.MethodPost).HandlerFunc(vb.RegisterOpenVPN)
	router.PathPrefix("/wg/register").Methods(http.MethodPost).HandlerFunc(vb.RegisterWireguard)
	router.PathPrefix("/openvpn/arrange ").Methods(http.MethodPost).HandlerFunc(vb.ArrangeOpenVPN)
	router.PathPrefix("/wg/arrange ").Methods(http.MethodPost).HandlerFunc(vb.ArrangeWireguard)

	return &vb
}

func appendBootstrapRouter(router *mux.Router, mongoURI string, pub *rsa.PublicKey) *serve.BootStrapBackend {
	logger.Info("Enable Bootstrap")
	cs, err := connstring.Parse(mongoURI)
	if err != nil {
		panic(err)
	}

	client, err := mongo.NewClient(
		options.Client().ApplyURI(mongoURI),
	)
	client.Connect(nil)
	if err != nil {
		panic(err)
	}

	backend := serve.BootStrapBackend{
		PublicKey: pub,
		DB:        client.Database(cs.Database),
	}
	router.PathPrefix("").Methods(http.MethodPost).HandlerFunc(backend.GenerateKey)

	return &backend
}
