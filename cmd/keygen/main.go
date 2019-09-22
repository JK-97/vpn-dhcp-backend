package main

import (
	"dhcp-backend/option"
	"dhcp-backend/serve"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

var appconfig *option.ApplicationConfig

func main() {
	appconfig = option.ReadConfigFile("keygen.cfg")
	mongoURI := appconfig.MongoURI

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
		PublicKey: appconfig.PublicKey,
		DB:        client.Database(cs.Database),
	}

	http.ListenAndServe(appconfig.Addr, http.HandlerFunc(backend.AddTicket))
}
