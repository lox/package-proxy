package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lox/package-proxy/cache"
	"github.com/lox/package-proxy/crypto"
	"github.com/lox/package-proxy/server"
	"github.com/lox/package-proxy/ubuntu"
)

const (
	listenHttp  = "0.0.0.0:3142"
	listenHttps = "0.0.0.0:3143"
	day         = time.Hour * 24
	week        = day * 7
	cacheSize   = 10 << 20 // 10Gb
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
	cache.NewPattern(`Release(\.gpg)?$`, time.Hour),
	cache.NewPattern(`Translation-(en|fr)\.(gz|bz2|bzip2|lzma)$`, time.Hour),
	cache.NewPattern(`Sources\.lzma$`, time.Hour),
}

type flags struct {
	EnableUbuntuRewriting bool
	EnableTlsUnwrapping   bool
	CacheDir              string
}

func parseFlags() flags {
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

	return flags{
		EnableUbuntuRewriting: *enableUbuntu,
		EnableTlsUnwrapping:   *enableTls,
		CacheDir:              *cacheDir,
	}
}

func enableTls(handler http.Handler) (http.Handler, error) {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	log.Printf("using certificate %s for dynamic cert generation",
		filepath.Join(path, "certs/public.pem"))

	// unwraps TLS with generated certificates
	handler, err = crypto.UnwrapTlsHandler(handler,
		"certs/private.key", "certs/public.pem")

	if err != nil {
		return nil, err
	}

	return handler, nil
}

func main() {
	flags := parseFlags()

	log.Printf("running package-proxy %s", version)

	cache, err := cache.NewDiskCache(flags.CacheDir, cacheSize)
	if err != nil {
		log.Fatal(err)
	}

	config := &server.Config{
		Cache:    cache,
		Patterns: cachePatterns,
	}

	if flags.EnableUbuntuRewriting {
		log.Printf("enabling ubuntu mirror rewriting")
		config.Rewriters = append(config.Rewriters, ubuntu.NewRewriter())
	}

	proxy, err := server.NewPackageProxy(config)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		log.Printf("http proxy listening on http://%s", listenHttp)
		log.Fatal(http.ListenAndServe(listenHttp, proxy))
	}()

	if flags.EnableTlsUnwrapping {
		wg.Add(1)

		handler, err := enableTls(proxy)
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			log.Printf("https proxy listening on https://%s", listenHttps)
			log.Fatal(http.ListenAndServe(listenHttps, handler))
		}()
	}

	wg.Wait()
}
