package providers

import "net/http"

type Provider interface {
	Match(r *http.Request) bool
	Rewrite(r *http.Request)
}

type defaultProvider struct {
}

func (p *defaultProvider) Match(r *http.Request) bool {
	return true
}

func (p *defaultProvider) Rewrite(r *http.Request) {
}

var DefaultProvider = &defaultProvider{}
