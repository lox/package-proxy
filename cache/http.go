package cache

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	ProxiedHeader  = "X-PackageProxy-Proxied"
	CachedHeader   = "X-PackageProxy-Cached"
	MaxAgeHeader   = "X-PackageProxy-MaxAge"
	CacheKeyHeader = "X-PackageProxy-Key"
)

func CachedRoundTripper(c Cache, upstream http.RoundTripper) *roundTripper {
	return &roundTripper{upstream: upstream, cache: c}
}

// roundTripper is a http.RoundTripper that caches responses
type roundTripper struct {
	upstream http.RoundTripper
	cache    Cache
}

func (r *roundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	key := req.Header.Get(CacheKeyHeader)

	// return immediately if we can't cache
	if key == "" {
		resp, err = r.upstream.RoundTrip(req)
		r.log(req, resp, "SKIP")
		return
	}

	cacheStatus := "HIT"

	// populate the cache if needed
	if !r.cache.Has(key) {
		cacheStatus = "MISS"
		upstreamResp, err := r.upstream.RoundTrip(req)
		if err != nil {
			return upstreamResp, err
		}

		if err := r.cache.WriteStream(key, upstreamResp.Body, false); err != nil {
			panic(err)
		}
	}

	stream, err := r.cache.ReadStream(key)
	if err != nil {
		return nil, err
	}

	// buffer into memory for Content-Length :(
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(stream)
	stream.Close()

	resp = new(http.Response)
	resp.Status = "200 OK"
	resp.StatusCode = 200
	resp.Proto = "HTTP/1.0"
	resp.ProtoMajor = 1
	resp.ProtoMinor = 0
	resp.Body = ioutil.NopCloser(buffer)
	resp.Request = req
	resp.ContentLength = int64(buffer.Len())
	resp.Header = http.Header{}
	resp.Header.Set(CachedHeader, "1")
	resp.Header.Set(CacheKeyHeader, key)

	r.log(req, resp, cacheStatus)

	return
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
