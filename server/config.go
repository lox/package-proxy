package server

import (
	"net/http"

	"github.com/lox/package-proxy/cache"
)

type Config struct {
	Rewriters []Rewriter
	Upstream  http.RoundTripper
	Cache     cache.Cache
	Patterns  cache.CachePatternSlice
}
