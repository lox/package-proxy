package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/lox/package-proxy/cache"
	"github.com/lox/package-proxy/server"
	"github.com/lox/package-proxy/ubuntu"
)

const (
	PROXY_ADDR = "0.0.0.0:3142"
	day        = time.Hour * 24
	week       = day * 7
)

var version string

var cachePatterns = cache.CachePatternSlice{
	// aptitude / ubuntu / debian
	cache.NewPattern(`deb$`, week),
	cache.NewPattern(`udeb$`, week),
	cache.NewPattern(`DiffIndex$`, time.Hour),
	cache.NewPattern(`PackagesIndex$`, time.Hour),
	cache.NewPattern(`Packages\.(bz2|gz|lzma)$`, time.Hour),
	cache.NewPattern(`SourcesIndex$`, time.Hour),
	cache.NewPattern(`Sources\.(bz2|gz|lzma)$`, time.Hour),
	cache.NewPattern(`Release$`, time.Hour),
	cache.NewPattern(`Translation-(en|fr)\.(gz|bz2|bzip2|lzma)$`, time.Hour),
	cache.NewPattern(`Sources\.lzma$`, time.Hour),
}

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: package-proxy [options]")
		fmt.Println("\nOptions:")
		fmt.Printf("  -dir=            The dir to store cache data in\n")
		fmt.Printf("  -tls=true        Enable tls and dynamic certificate generation\n")
		fmt.Printf("  -ubuntu=true     Rewrite ubuntu urls to the fastest mirror\n")
	}

	cacheDir := flag.String("dir", "", "The dir to store cache data in")
	enableTls := flag.Bool("tls", false, "Enable tls and dynamic certificate generation")
	enableUbuntu := flag.Bool("ubuntu", true, "Rewrite ubuntu urls to the fastest mirror")
	flag.Parse()

	log.Printf("running package-proxy %s", version)

	// disk-backed cache, 10GB
	cache, err := cache.NewDiskCache(*cacheDir, 10<<20)
	if err != nil {
		log.Fatal(err)
	}

	config := &server.Config{
		EnableTls: *enableTls,
		Cache:     cache,
		Patterns:  cachePatterns,
	}

	if *enableUbuntu {
		log.Printf("enabling ubuntu mirror rewriting")
		config.Rewriters = append(config.Rewriters, ubuntu.NewRewriter())
	}

	proxy, err := server.NewPackageProxy(config)
	if err != nil {
		log.Fatal(err)
	}

	protocol := "http"
	if *enableTls {
		protocol = "https"
	}

	log.Printf("proxy listening on %s://%s", protocol, PROXY_ADDR)
	log.Fatal(http.ListenAndServe(PROXY_ADDR, proxy))
}
