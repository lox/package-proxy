package server

import (
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/lox/packageproxy/cache"
	"github.com/lox/packageproxy/crypto"
)

type PackageCache struct {
	Handler http.Handler
	Config  Config
}

type RequestRewriter interface {
	Rewrite(req *http.Request)
}

func NewPackageCache(config Config) (*PackageCache, error) {
	proxy := &httputil.ReverseProxy{
		Transport: cache.CachedRoundTripper(cache.NewMemoryCache(), &http.Transport{}),
	}

	handler, err := crypto.UnwrapTlsHandler(proxy, "certs/private.key", "certs/public.pem")
	if err != nil {
		return nil, err
	}

	pm := &PackageCache{Handler: handler, Config: config}
	proxy.Director = pm.director

	return pm, nil
}

func (p *PackageCache) director(req *http.Request) {
	for _, provider := range p.Config.Providers {
		provider.Rewrite(req)
	}
}

func (p *PackageCache) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var matched bool

	for _, provider := range p.Config.Providers {
		if provider.Match(req) {
			log.Printf("request matched provider %#v", provider)
			matched = true
			break
		}
	}

	if !matched {
		//TODO: error
	}

	p.Handler.ServeHTTP(rw, req)
}
