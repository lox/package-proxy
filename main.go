package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lox/package-proxy/cache"
	"github.com/lox/package-proxy/mitm"
	"github.com/lox/package-proxy/server"
	"github.com/lox/package-proxy/ubuntu"
	"github.com/nu7hatch/gouuid"
)

const (
	listen    = "0.0.0.0:3142"
	day       = time.Hour * 24
	week      = day * 7
	forever   = day * 1000
	cacheSize = 10 << 20 // 10Gb
	caKey     = "certs/packageproxy-ca.key"
	caCert    = "certs/packageproxy-ca.crt"
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
	// composer / packagist
	cache.NewPattern(`^https?://packagist\.org/(.+)\.json$`, time.Hour),
	cache.NewPattern(`^https://api.github.com/repos/Seldaek/jsonlint/zipball/1.0.0`, week),
	// github
	cache.NewPattern(`^https://codeload.github.com/(.+)/legacy.zip/(.+)$`, week),
	cache.NewPattern(`^https://api.github.com/repos/(.+)/zipball`, week),
	// bitbucket
	cache.NewPattern(`^https://bitbucket.org/(.+).zip$`, week),
	// rubygems
	cache.NewPattern(`/api/v1/dependencies`, day),
	cache.NewPattern(`gem\$`, week),
	// npm
	cache.NewPattern(`^https?://cnpmjs.org/(.+)\.tgz$`, week),
	cache.NewPattern(`^https?://registry.npmjs.org/(.+)\.tgz$`, week),
	cache.NewPattern(`^https?://registry.npmjs.org/`, time.Hour),
}

var mitmHosts = []string{
	"codeload.github.com:443",
	"registry.npmjs.org:443",
	"api.github.com:443",
	"packagist.org:443",
}

type flags struct {
	EnableRewrites      []string
	EnableTlsUnwrapping bool
	CacheDir            string
	ShowVersion         bool
}

func parseFlags() flags {
	flag.Usage = func() {
		fmt.Println("Usage: package-proxy [options]")
		fmt.Println("\nOptions:")
		fmt.Printf("  -dir=            The dir to store cache data in\n")
		fmt.Printf("  -tls=true        Enable tls and dynamic certificate generation\n")
		fmt.Printf("  -rewrite=all     Only rewrite specific services (defaults to all)\n")
		fmt.Printf("  -version         The compiled version\n")
	}

	cacheDir := flag.String("dir", "", "The dir to store cache data in")
	enableTls := flag.Bool("tls", false, "Enable tls and dynamic certificate generation")
	enableRewrites := flag.String("rewrite", "all", "Only rewrite specific services")
	showVersion := flag.Bool("version", false, "Show the compiled version")
	flag.Parse()

	return flags{
		EnableRewrites:      strings.Split(*enableRewrites, ","),
		EnableTlsUnwrapping: *enableTls,
		CacheDir:            *cacheDir,
		ShowVersion:         *showVersion,
	}
}

func enableTls(handler http.Handler) (http.Handler, error) {
	log.Printf("using ca cert %s for tls unwrapping", caCert)
	mitmHandler, err := mitm.InterceptTlsHandler(handler, caKey, caCert)
	if err != nil {
		return handler, err
	}

	for _, host := range mitmHosts {
		mitmHandler.AddHost(host)
	}

	return mitmHandler, nil
}

func isRewriterEnabled(service string, enabled []string) bool {
	for _, e := range enabled {
		if e == "all" || e == service {
			return true
		}
	}

	return false
}

func buildRewriters(enabled []string) []server.Rewriter {
	r := []server.Rewriter{}

	if isRewriterEnabled("ubuntu", enabled) {
		log.Printf("enabling ubuntu mirror rewriting")
		r = append(r, ubuntu.NewRewriter())
	}
	return r
}

func main() {
	flags := parseFlags()

	if flags.ShowVersion {
		fmt.Printf("package-proxy %s\n", version)
		os.Exit(0)
	}

	log.Printf("running package-proxy %s", version)

	cache, err := cache.NewDiskCache(flags.CacheDir, cacheSize)
	if err != nil {
		log.Fatal(err)
	}

	uid, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("server id is %s", uid.String())

	config := &server.Config{
		Cache:     cache,
		Patterns:  cachePatterns,
		Rewriters: buildRewriters(flags.EnableRewrites),
		ServerId:  uid.String(),
	}

	if version != "" {
		config.ServerId += " (package-proxy/" + version + ")"
	} else {
		config.ServerId += " (package-proxy)"
	}

	var handler http.Handler

	handler, err = server.NewPackageProxy(config)
	if err != nil {
		log.Fatal(err)
	}

	if flags.EnableTlsUnwrapping {
		handler, err = enableTls(handler)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("proxy listening on https://%s", listen)
	log.Fatal(http.ListenAndServe(listen, handler))
}
