//+build ignore

// generates id_rsa id_rsa.go

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func main() {
	pri, err := rsa.GenerateKey(rand.Reader, 1024)

	if err != nil {
		panic(err)
	}
	pri.Precompute()

	pub := pri.PublicKey
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(pri),
	}

	f, err := os.Create("id_rsa")
	if err != nil {
		panic(err)
	}
	pem.Encode(f, block)

	block = &pem.Block{
		Type: "PUBLIC KEY",

		Bytes: x509.MarshalPKCS1PublicKey(&pub),
	}
	f, err = os.Create("id_rsa.pub")
	if err != nil {
		panic(err)
	}
	pem.Encode(f, block)
}
