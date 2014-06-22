package cache

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"path/filepath"
	"time"
)

const (
	MaxAgeHeader       = "X-Package-Proxy-MaxAge"
	CacheHeader        = "X-Package-Proxy"
	CanonicalUrlHeader = "X-Canonical-Url"
)

// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html#sec13.5.1
var ignoredHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"TE",
	"Trailers",
	"Transfer - Encoding",
	"Upgrade",
}

func CachedRoundTripper(c Cache, upstream http.RoundTripper) *roundTripper {
	return &roundTripper{upstream: upstream, cache: c}
}

// roundTripper is a http.RoundTripper that caches responses
type roundTripper struct {
	upstream http.RoundTripper
	cache    Cache
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {

	// return immediately if we can't cache
	if !isRequestCacheable(req) {
		resp, err := r.upstream.RoundTrip(req)
		if err != nil {
			log.Println(err)
			return resp, err
		}

		resp.Header.Set(CacheHeader, "SKIP")
		logRequest(req, resp, "SKIP")
		return resp, nil
	}

	key := cacheKey(req)
	cacheStatus := "HIT"

	// populate the cache if needed
	if !r.cache.Has(key) {
		cacheStatus = "MISS"

		upstreamResp, err := r.upstream.RoundTrip(req)
		if err != nil {
			return upstreamResp, err
		}

		if !isResponseCacheable(upstreamResp) {
			logRequest(req, upstreamResp, "SKIP")
			return upstreamResp, nil
		}

		var maxAge time.Duration

		// TODO: check response max-age
		if m := req.Header.Get(MaxAgeHeader); m != "" {
			maxAge, err = time.ParseDuration(m)
			if err != nil {
				log.Println(err)
			}
		}

		// strip unwanted response headers
		for _, h := range ignoredHeaders {
			upstreamResp.Header.Del(h)
		}

		b, err := httputil.DumpResponse(upstreamResp, true)
		if err != nil {
			panic(err)
		}

		defer upstreamResp.Body.Close()
		err = r.cache.Write(key, bytes.NewReader(b), maxAge)
		if err != nil {
			panic(err)
		}
	}

	stream, err := r.cache.Read(key)
	if err != nil {
		return nil, err
	}

	b := bufio.NewReader(stream)
	defer stream.Close()

	resp, err := http.ReadResponse(b, req)
	if err != nil {
		panic(err)
	}

	if age, err := calculateAge(resp); err != nil {
		log.Println("unable to parse date header", err)
	} else {
		resp.Header.Set("Age", fmt.Sprintf("%.f", age.Seconds()))
	}

	resp.Header.Set(CacheHeader, fmt.Sprintf("%s", cacheStatus))

	logRequest(req, resp, cacheStatus)
	return resp, err
}

func calculateAge(resp *http.Response) (d time.Duration, err error) {
	stored, err := http.ParseTime(resp.Header.Get("Date"))
	if err != nil {
		return d, err
	}

	return time.Now().Sub(stored), nil
}

func isRequestCacheable(req *http.Request) bool {
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

func isResponseCacheable(resp *http.Response) bool {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	} else if resp.StatusCode == http.StatusFound {
		return true
	} else {
		return false
	}
}

func cacheKey(req *http.Request) string {
	url := req.URL

	if h := req.Header.Get(CanonicalUrlHeader); h != "" {
		if u, err := url.Parse(h); err == nil {
			url = u
		} else {
			log.Printf("ignoring invalid canonical url %s", h)
		}
	}

	return filepath.Join(url.Host, url.Path+"/"+req.Method)
}

func logRequest(req *http.Request, resp *http.Response, status string) {
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
