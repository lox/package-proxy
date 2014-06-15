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

func CachedRoundTripper(c Cache, rt http.RoundTripper) *cachedRoundTripper {
	return &cachedRoundTripper{RoundTripper: rt, cache: c}
}

// cachedRoundTripper is a http.RoundTripper that caches responses
type cachedRoundTripper struct {
	http.RoundTripper
	cache Cache
}

func (r *cachedRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	cachedResp, err := CachedResponse(r.cache, req)
	if err != nil {
		panic(err)
	}

	if cachedResp != nil {
		cachedResp.Header.Set(CachedHeader, "1")
		resp = cachedResp
	} else {
		resp, err = r.RoundTripper.RoundTrip(req)
		if err == nil && resp.StatusCode == 200 {
			if b, derr := httputil.DumpResponse(resp, true); derr == nil {
				go r.cache.Set(RequestCacheKey(req), b)
			}
		}
	}

	resp.Header.Set(ProxiedHeader, "1")
	r.log(req, resp)

	return
}

func (r *cachedRoundTripper) log(req *http.Request, resp *http.Response) {
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
