package server

import "github.com/lox/package-proxy/providers"

type Config struct {
	Providers    []providers.Provider
	CacheDir     string
	CacheSizeMax uint64
}
