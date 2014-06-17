package server

import "github.com/lox/package-proxy/providers"

type Config struct {
	EnableTls    bool
	Providers    []providers.Provider
	CacheDir     string
	CacheSizeMax uint64
}
