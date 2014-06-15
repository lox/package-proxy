package main

import (
	"log"
	"net/http"

	"github.com/lox/packageproxy/providers"
	"github.com/lox/packageproxy/providers/ubuntu"
	"github.com/lox/packageproxy/server"
)

const (
	PROXY_ADDR = "0.0.0.0:8080"
)

func main() {
	config := server.Config{
		Providers: []providers.Provider{
			ubuntu.NewProvider(),
			providers.DefaultProvider,
		},
	}

	pc, err := server.NewPackageCache(config)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("proxy listening on", PROXY_ADDR)
	log.Fatal(http.ListenAndServe(PROXY_ADDR, pc))
}
