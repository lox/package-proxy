package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	pmcproxy "github.com/lox/pmc"
)

type testFixture struct {
	proxy, backend *httptest.Server
}

func newTestFixture(handler http.HandlerFunc) *testFixture {
	pmc, err := NewPmc()
	if err != nil {
		panic(err)
	}

	backend := httptest.NewServer(http.HandlerFunc(handler))
	backendUrl, err := url.Parse(backend.URL)
	if err != nil {
		panic(err)
	}

	// force rewriting to the testing backend
	pmc.Proxy.Director = func(req *http.Request) {
		req.URL.Scheme = backendUrl.Scheme
		req.URL.Host = backendUrl.Host
		log.Printf("%#v", req)
	}

	return &testFixture{httptest.NewServer(pmc), backend}
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

	assertHeader(t, resp1, pmcproxy.ProxiedHeader, "1")
	assertHeader(t, resp1, pmcproxy.CachedHeader, "")

	// second request should be cached
	resp2, err := fixture.client().Get(fixture.backend.URL)
	if err != nil {
		t.Fatal(err)
	}

	assertHeader(t, resp2, pmcproxy.ProxiedHeader, "1")
	assertHeader(t, resp2, pmcproxy.CachedHeader, "1")
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

	assertHeader(t, resp1, pmcproxy.ProxiedHeader, "1")
	assertHeader(t, resp1, pmcproxy.CachedHeader, "")
}
