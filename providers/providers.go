package providers

import (
	"net/http"
	"net/url"
	"time"

	"github.com/lox/package-proxy/cache"
)

type Provider interface {
	Match(r *http.Request) bool
	RewriteUrl(u *url.URL)
	CacheKey(r *http.Request) string
	CacheMaxAge(r *http.Request) time.Duration
}

type roundTripper struct {
	http.RoundTripper
	providers []Provider
}

// ProviderRoundTripper matches a provider, sets some headers and delegates to upstream
func ProviderRoundTripper(providers []Provider, upstream http.RoundTripper) *roundTripper {
	return &roundTripper{RoundTripper: upstream, providers: providers}
}

func (r *roundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	var provider Provider

	// find a provider that matches
	for _, p := range r.providers {
		if p.Match(req) {
			provider = p
			break
		}
	}

	// apply caching and rewriting if a provider is found
	if provider != nil {
		req.Header.Set(cache.CacheKeyHeader, provider.CacheKey(req))
		req.Header.Set(cache.MaxAgeHeader, provider.CacheMaxAge(req).String())
		provider.RewriteUrl(req.URL)
	}

	return r.RoundTripper.RoundTrip(req)
}
