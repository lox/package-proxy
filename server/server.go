package server

import (
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"

	"github.com/lox/package-proxy/cache"
	"github.com/lox/package-proxy/crypto"
	"github.com/lox/package-proxy/providers"
	"github.com/peterbourgon/diskv"
)

type PackageProxy struct {
	http.Handler
	Config Config
}

type RequestRewriter interface {
	Rewrite(req *http.Request)
}

func NewPackageProxy(config Config) (*PackageProxy, error) {

	// basic disk-backed cache, keys are filenames
	diskCache := diskv.New(diskv.Options{
		BasePath: config.CacheDir,
		Transform: func(key string) []string {
			path := filepath.Join(config.CacheDir, key)
			os.MkdirAll(filepath.Dir(path), 0700)
			return []string{}
		},
		CacheSizeMax: config.CacheSizeMax,
	})

	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {},
		Transport: providers.ProviderRoundTripper(config.Providers,
			cache.CachedRoundTripper(diskCache, &http.Transport{}),
		),
	}

	// unwraps TLS with generated certificates
	handler, err := crypto.UnwrapTlsHandler(proxy, "certs/private.key", "certs/public.pem")
	if err != nil {
		return nil, err
	}

	return &PackageProxy{Handler: handler, Config: config}, nil
}

func (p *PackageProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Printf("Proxying %s %s", req.Method, req.URL.String())
	p.Handler.ServeHTTP(rw, req)
}
