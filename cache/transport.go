package cache

import (
	"log"
	"net/http"
	"net/http/httputil"
)

const (
	ProxiedHeader = "X-Pmc-Proxy"
	CachedHeader  = "X-Pmc-Cached"
)

// RoundTripper
type RoundTripper struct {
	http.RoundTripper
	Cache Cache
}

func (r *RoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	cachedResp, err := CachedResponse(r.Cache, req)
	if err != nil {
		panic(err)
	}

	if cachedResp != nil {
		cachedResp.Header.Set(CachedHeader, "1")
		resp = cachedResp
	} else {
		resp, err = r.RoundTripper.RoundTrip(req)
		if err == nil {
			if b, derr := httputil.DumpResponse(resp, true); derr == nil {
				go r.cacheResponse(req, b)
			}
		}
	}

	resp.Header.Set(ProxiedHeader, "1")
	r.log(req, resp)

	return
}

func (r *RoundTripper) cacheResponse(req *http.Request, bytes []byte) {
	r.Cache.Set(RequestCacheKey(req), bytes)
}

func (r *RoundTripper) log(req *http.Request, resp *http.Response) {
	var status string
	switch resp.Header.Get(CachedHeader) {
	case "1":
		status = "HIT"
	default:
		status = "MISS"
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
