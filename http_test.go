package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/lox/package-proxy/cache"
	"github.com/lox/package-proxy/providers"
	"github.com/lox/package-proxy/server"
)

type testFixture struct {
	proxy, backend *httptest.Server
	provider       *testProvider
}

type testProvider struct {
	matcherFunc func(r *http.Request) bool
	rewriteFunc func(r *http.Request)
}

func (p *testProvider) Match(r *http.Request) bool {
	return p.matcherFunc(r)
}

func (p *testProvider) Rewrite(r *http.Request) {
	p.rewriteFunc(r)
}

func newTestFixture(handler http.HandlerFunc) *testFixture {
	provider := &testProvider{}
	pp, err := server.NewPackageProxy(server.Config{Providers: []providers.Provider{
		provider,
	}})
	if err != nil {
		panic(err)
	}

	backend := httptest.NewServer(http.HandlerFunc(handler))
	backendUrl, err := url.Parse(backend.URL)
	if err != nil {
		panic(err)
	}

	provider.matcherFunc = func(r *http.Request) bool {
		return true
	}
	provider.rewriteFunc = func(r *http.Request) {
		r.URL.Host = backendUrl.Host
	}

	return &testFixture{httptest.NewServer(pp), backend, provider}
}

// client returns an http client configured to use the provided proxy
func (f *testFixture) client() *http.Client {
	proxyUrl, err := url.Parse(f.proxy.URL)
	if err != nil {
		panic(err)
	}

	return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
}

func (f *testFixture) close() {
	f.proxy.Close()
	f.backend.Close()
}

func assertHeader(t *testing.T, r *http.Response, header string, expected string) {
	if r.Header.Get(header) != expected {
		t.Fatalf("Expected header %s=1, but got '%s'", header, r.Header.Get(header))
	}
}

func TestProxyCachesRequests(t *testing.T) {
	fixture := newTestFixture(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Llamas rock"))
	})
	defer fixture.close()

	// first request shouldn't be cached
	resp1, err := fixture.client().Get(fixture.backend.URL)
	if err != nil {
		t.Fatal(err)
	}

	assertHeader(t, resp1, cache.ProxiedHeader, "1")
	assertHeader(t, resp1, cache.CachedHeader, "")

	// second request should be cached
	resp2, err := fixture.client().Get(fixture.backend.URL)
	if err != nil {
		t.Fatal(err)
	}

	assertHeader(t, resp2, cache.ProxiedHeader, "1")
	assertHeader(t, resp2, cache.CachedHeader, "1")
}

func TestUbuntuRewrites(t *testing.T) {
	fixture := newTestFixture(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Llamas rock"))
	})
	defer fixture.close()

	resp1, err := fixture.client().Get("http://archive.ubuntu.com/ubuntu/pool/main/b/bind9/dnsutils_9.9.5.dfsg-3_amd64.deb")
	if err != nil {
		t.Fatal(err)
	}

	assertHeader(t, resp1, cache.ProxiedHeader, "1")
	assertHeader(t, resp1, cache.CachedHeader, "")
}
