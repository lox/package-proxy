package cache

import (
	"bufio"
	"bytes"
	"log"
	"net/http"
	"net/http/httputil"
	"path/filepath"
)

const (
	MaxAgeHeader = "X-PackageProxy-MaxAge"
)

func CachedRoundTripper(c Cache, upstream http.RoundTripper) *roundTripper {
	return &roundTripper{upstream: upstream, cache: c}
}

// roundTripper is a http.RoundTripper that caches responses
type roundTripper struct {
	upstream http.RoundTripper
	cache    Cache
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// TODO: respect max-age
	// maxAge := req.Header.Get(MaxAgeHeader)

	// return immediately if we can't cache
	if !r.isRequestCacheable(req) {
		resp, err := r.upstream.RoundTrip(req)
		r.log(req, resp, "SKIP")
		return resp, err
	}

	key := r.cacheKey(req)
	cacheStatus := "HIT"

	// populate the cache if needed
	if !r.cache.Has(key) {
		cacheStatus = "MISS"

		upstreamResp, err := r.upstream.RoundTrip(req)
		if err != nil {
			return upstreamResp, err
		}

		// only cache GET 2xx statuses
		if upstreamResp.StatusCode < 200 || upstreamResp.StatusCode > 200 {
			r.log(req, upstreamResp, "SKIP")
			return upstreamResp, nil
		}

		b, err := httputil.DumpResponse(upstreamResp, true)
		if err != nil {
			panic(err)
		}

		defer upstreamResp.Body.Close()
		err = r.cache.WriteStream(key, bytes.NewReader(b), false)
		if err != nil {
			panic(err)
		}
	}

	stream, err := r.cache.ReadStream(key)
	if err != nil {
		return nil, err
	}

	b := bufio.NewReader(stream)
	defer stream.Close()

	resp, err := http.ReadResponse(b, req)
	if err != nil {
		panic(err)
	}

	r.log(req, resp, cacheStatus)
	return resp, err
}

func (r *roundTripper) isRequestCacheable(req *http.Request) bool {
	if req.Header.Get(MaxAgeHeader) == "" {
		return false
	}

	switch req.Method {
	case "GET":
		return true
	case "HEAD":
		return true
	}

	return false
}

func (r *roundTripper) cacheKey(req *http.Request) string {
	return filepath.Join(req.URL.Host, req.URL.Path+"/"+req.Method)
}

func (r *roundTripper) log(req *http.Request, resp *http.Response, status string) {
	if status == "HIT" {
		status = "\x1b[32;1m" + status + "\x1b[0m"
	} else if status == "MISS" {
		status = "\x1b[31;1m" + status + "\x1b[0m"
	}
	log.Printf(
		"%s \"%s %s %s\" (%s) %d %s",
		req.RemoteAddr,
		req.Method,
		req.URL,
		req.Proto,
		resp.Status,
		resp.ContentLength,
		status,
	)
}
