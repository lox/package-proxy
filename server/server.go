package server

import (
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"

	"github.com/lox/package-proxy/cache"
	"github.com/lox/package-proxy/crypto"
)

type PackageProxy struct {
	Handler   http.Handler
	Cache     cache.Cache
	Transport *http.Transport
	Rewriters []Rewriter
	Patterns  cache.CachePatternSlice
}

type Rewriter interface {
	Rewrite(req *http.Request)
}

type RewriterFunc func(req *http.Request)

// Rewrite calls f(req).
func (f RewriterFunc) Rewrite(req *http.Request) {
	f(req)
}

func applyConfigDefaults(config *Config) error {
	if config.Upstream == nil {
		config.Upstream = &http.Transport{}
	}

	if config.Patterns == nil {
		config.Patterns = cache.CachePatternSlice{}
	}

	if config.Cache == nil {
		cache, err := cache.NewDiskCache("", 1<<20)
		if err != nil {
			return err
		}
		config.Cache = cache
	}

	return nil
}

func NewPackageProxy(config *Config) (*PackageProxy, error) {
	if err := applyConfigDefaults(config); err != nil {
		return nil, err
	}

	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			// these get applied to the upstream request
			for _, rewrite := range config.Rewriters {
				rewrite.Rewrite(r)
			}
		},
		Transport: cache.CachedRoundTripper(config.Cache, config.Upstream),
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

	return &PackageProxy{
		Handler:   handler,
		Cache:     config.Cache,
		Rewriters: config.Rewriters,
		Patterns:  config.Patterns,
	}, nil
}

func (p *PackageProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	match, pattern := p.Patterns.MatchString(req.URL.String())
	if match {
		req.Header.Set(cache.MaxAgeHeader, pattern.Duration.String())
	}

	req.Header.Set("X-Canonical-Url", req.URL.String())
	p.Handler.ServeHTTP(rw, req)
}
