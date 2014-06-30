package server

import (
	"net/http"
	"net/http/httputil"

	"github.com/lox/package-proxy/cache"
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

	if config.ServerId == "" {
		config.ServerId = "package-proxy"
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
			// strip conditional request headers
			r.Header.Del("If-Modified-Since")
			r.Header.Del("If-Match")

			// these get applied to the upstream request
			for _, rewrite := range config.Rewriters {
				rewrite.Rewrite(r)
			}

			// reset host header
			r.Host = r.URL.Host
		},
		Transport: cache.CachedRoundTripper(
			config.Cache, config.Upstream, config.ServerId,
		),
	}

	var handler http.Handler
	handler = proxy

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
