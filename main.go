package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/lox/package-proxy/providers"
	"github.com/lox/package-proxy/providers/ubuntu"
	"github.com/lox/package-proxy/server"
)

const (
	PROXY_ADDR = "0.0.0.0:3142"
)

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: package-proxy [options]")
		fmt.Println("\nOptions:")
		fmt.Printf("  -dir=.cache    The dir to store cache data in\n")
	}

	cacheDir := flag.String("dir", ".cache", "The dir to store cache data in")
	flag.Parse()

	log.Printf("storing cache in %s", *cacheDir)

	config := server.Config{
		CacheDir:     *cacheDir,
		CacheSizeMax: 1024 << 20, // 1GB
		Providers: []providers.Provider{
			ubuntu.NewProvider(),
		},
	}

	pc, err := server.NewPackageProxy(config)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("proxy listening on", PROXY_ADDR)
	log.Fatal(http.ListenAndServe(PROXY_ADDR, pc))
}
