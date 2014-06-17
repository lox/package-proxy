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
		fmt.Printf("  -tls=true      Enable tls and dynamic certificate generation\n")
	}

	cacheDir := flag.String("dir", ".cache", "The dir to store cache data in")
	enableTls := flag.Bool("tls", false, "Enable tls and dynamic certificate generation")
	flag.Parse()

	log.Printf("storing cache in %s", *cacheDir)

	config := server.Config{
		EnableTls:    *enableTls,
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

	protocol := "http"
	if config.EnableTls {
		protocol = "https"
	}

	log.Printf("proxy listening on %s://%s", protocol, PROXY_ADDR)
	log.Fatal(http.ListenAndServe(PROXY_ADDR, pc))
}
