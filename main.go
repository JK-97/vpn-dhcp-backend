package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"
)

func main() {
	kvc, err := clientv3.New(clientv3.Config{
		Endpoints: []string{
			"10.55.2.114:2379",
		}})

	if err != nil {
		panic(err)
	}

	agent := EtcdDNSAgent{
		cli:    kvc,
		Prefix: "/skydns",
	}
	resp, _ := kvc.Get(context.Background(), agent.DomainToKey("jx.skydns.local"), clientv3.WithPrefix())

	fmt.Println(resp)

	backend := DNSBackend{Agent: &agent}

	r := mux.NewRouter()

	r.NotFoundHandler = NewNotFoundHandler()
	r.Use(simpleLoggingMw)

	router := r.PathPrefix("/api/v1/dns").Subrouter()
	router.PathPrefix("").Methods(http.MethodPost).HandlerFunc(backend.AddRecord)
	router.PathPrefix("").Methods(http.MethodDelete).HandlerFunc(backend.RemoveRecord)
	router.PathPrefix("").Methods(http.MethodGet).HandlerFunc(backend.FindRecord)

	http.ListenAndServe(":1054", r)
}
