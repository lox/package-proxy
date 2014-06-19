package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/lox/package-proxy/server"
	"github.com/lox/package-proxy/ubuntu"
)

const (
	PROXY_ADDR = "0.0.0.0:3142"
)

var version string

var Rewriters []server.Rewriter

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: package-proxy [options]")
		fmt.Println("\nOptions:")
		fmt.Printf("  -dir=          The dir to store cache data in\n")
		fmt.Printf("  -tls=true      Enable tls and dynamic certificate generation\n")
		fmt.Printf("  -ubuntu=true   Rewrite ubuntu urls to the fastest mirror\n")
	}

	cacheDir := flag.String("dir", "", "The dir to store cache data in")
	enableTls := flag.Bool("tls", false, "Enable tls and dynamic certificate generation")
	enableUbuntu := flag.Bool("ubuntu", true, "Rewrite ubuntu urls to the fastest mirror")
	flag.Parse()

	// default to a system temp directory
	if *cacheDir == "" {
		tmpDir := filepath.Join(os.TempDir(), "package-proxy/default")
		cacheDir = &tmpDir
	}

	log.Printf("running package-proxy %s", version)
	log.Printf("storing cache in %s", *cacheDir)

	if *enableUbuntu {
		log.Printf("enabling ubuntu mirror rewriting")
		Rewriters = append(Rewriters, ubuntu.NewRewriter())
	}

	proxy, err := server.NewPackageProxy(server.Config{
		EnableTls: *enableTls,
		CacheDir:  *cacheDir,
		Rewriters: Rewriters,
	})
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
