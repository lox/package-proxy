package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/lox/package-proxy/cache"
	"github.com/lox/package-proxy/server"
)

type testFixture struct {
	proxy, backend *httptest.Server
}

func newTestFixture(handler http.HandlerFunc, conf *server.Config) *testFixture {
	backend := httptest.NewServer(http.HandlerFunc(handler))
	backendUrl, err := url.Parse(backend.URL)
	if err != nil {
		panic(err)
	}

	if conf.Rewriters == nil {
		conf.Rewriters = []server.Rewriter{}
	}

	if conf.Cache == nil {
		conf.Cache = cache.NewMapCache()
	}

	// force all requests to rewrite to the test server
	conf.Rewriters = append(conf.Rewriters, server.RewriterFunc(func(req *http.Request) {
		req.URL.Host = backendUrl.Host
	}))

	pp, err := server.NewPackageProxy(conf)
	if err != nil {
		panic(err)
	}

	return &testFixture{httptest.NewServer(pp), backend}
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
		t.Fatalf("Expected header %s=%s, but got '%s'",
			header, expected, r.Header.Get(header))
	}
}

func TestProxyCachesRequests(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Llamas rock"))
	}
	fixture := newTestFixture(handler, &server.Config{
		Patterns: cache.CachePatternSlice{
			cache.NewPattern(".", time.Hour*100),
		},
	})
	defer fixture.close()

	// first request shouldn't be cached
	resp1, err := fixture.client().Get(fixture.backend.URL)
	if err != nil {
		t.Fatal(err)
	}

	assertHeader(t, resp1, cache.CacheHeader, "MISS")

	// second request should be cached
	resp2, err := fixture.client().Get(fixture.backend.URL)
	if err != nil {
		t.Fatal(err)
	}

	assertHeader(t, resp2, cache.CacheHeader, "HIT")
}

func TestRewritesApply(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Llamas rock"))
	}
	fixture := newTestFixture(handler, &server.Config{
		Patterns: cache.CachePatternSlice{
			cache.NewPattern(".", time.Hour*100),
		},
	})
	defer fixture.close()

	resp, err := fixture.client().Get(
		"http://archive.ubuntu.com/ubuntu/pool/main/b/bind9/dnsutils_9.9.5.dfsg-3_amd64.deb")
	if err != nil {
		t.Fatal(err)
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	assertHeader(t, resp, cache.CacheHeader, "MISS")

	if bytes.Compare(contents, []byte("Llamas rock")) != 0 {
		t.Fatalf("Response content was incorrect, rewrites not applying?")
	}
}
