package server

import (
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"time"

	"github.com/lox/package-proxy/cache"
	"github.com/lox/package-proxy/crypto"
)

var cachePatterns = cache.CachePatternSlice{
	cache.NewPattern(`deb$`, time.Hour*1000),
	cache.NewPattern(`udeb$`, time.Hour*1000),
	cache.NewPattern(`DiffIndex$`, time.Hour*24),
	cache.NewPattern(`PackagesIndex$`, time.Hour*24),
	cache.NewPattern(`Packages\.(bz2|gz|lzma)$`, time.Hour*24),
	cache.NewPattern(`SourcesIndex$`, time.Hour*24),
	cache.NewPattern(`Sources\.(bz2|gz|lzma)$`, time.Hour*24),
	cache.NewPattern(`Release$`, time.Hour*24),
	cache.NewPattern(`Translation-(en|fr)\.(gz|bz2|bzip2|lzma)$`, time.Hour*24),
	cache.NewPattern(`Sources\.lzma$`, time.Hour*24),
}

type PackageProxy struct {
	http.Handler
	Config Config
}

type Rewriter interface {
	Rewrite(req *http.Request)
}

type RewriterFunc func(req *http.Request)

func NewPackageProxy(config Config) (*PackageProxy, error) {

	diskCache, err := cache.NewFileCache(config.CacheDir)
	if err != nil {
		return nil, err
	}

	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
		},
		Transport: cache.CachedRoundTripper(diskCache, &http.Transport{}),
	}

	var handler http.Handler
	handler = proxy

	if config.EnableTls {
		path, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		log.Printf("using certificate %s for dynamic cert generation",
			filepath.Join(path, "certs/public.pem"))

		// unwraps TLS with generated certificates
		handler, err = crypto.UnwrapTlsHandler(proxy, "certs/private.key", "certs/public.pem")
		if err != nil {
			return nil, err
		}
	}

	return &PackageProxy{Handler: handler, Config: config}, nil
}

func (p *PackageProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	match, pattern := cachePatterns.MatchString(req.URL.String())
	if match {
		req.Header.Set(cache.MaxAgeHeader, pattern.Duration.String())
	}

	p.Handler.ServeHTTP(rw, req)
}
