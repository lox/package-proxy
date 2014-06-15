package server

import "github.com/lox/packageproxy/providers"

type Config struct {
	Providers []providers.Provider
}
