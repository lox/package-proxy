package cache

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

const (
	MaxAgeHeader       = "X-Package-Proxy-MaxAge"
	CacheHeader        = "X-Cache"
	CacheLookupHeader  = "X-Cache-Lookup"
	CanonicalUrlHeader = "X-Canonical-Url"
)

// CachedRoundTripper either uses a cache for serving a response, or the provided upstream
func CachedRoundTripper(c Cache, upstream http.RoundTripper, serverId string) *roundTripper {
	return &roundTripper{
		upstream: upstream,
		cache:    c,
		serverId: serverId,
	}
}

// roundTripper is a http.RoundTripper that caches responses
type roundTripper struct {
	upstream http.RoundTripper
	cache    Cache
	serverId string
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	key := cacheKey(req)

	if isRequestCacheable(req) && r.cache.Has(key) {
		stream, err := r.cache.Read(key)
		if err != nil {
			return nil, err
		}

		defer stream.Close()
		resp, err := http.ReadResponse(bufio.NewReader(stream), req)
		if err != nil {
			return resp, err
		}

		return r.cacheHit(resp)
	}

	upstreamResp, err := r.upstream.RoundTrip(req)
	if err != nil {
		return upstreamResp, err
	}

	if isRequestCacheable(req) && isResponseCacheable(upstreamResp) {
		respCopy, err := copyResponse(upstreamResp)
		if err != nil {
			return nil, err
		}
		go r.storeResponse(key, respCopy)
	} else {
		return r.cacheSkip(upstreamResp)
	}

	return r.cacheMiss(upstreamResp)
}

// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html#sec13.5.1
var ignoredHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"TE",
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
	"X-Cache",
	"X-Cache-Lookup",
	"Via",
}

// storeResponse writes a response to the cache
func (r *roundTripper) storeResponse(key string, resp *http.Response) error {
	for _, h := range ignoredHeaders {
		resp.Header.Del(h)
	}

	b, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// TODO: check Response Cache-Control headers
	if header := resp.Request.Header.Get(MaxAgeHeader); header != "" {
		if maxAge, err := time.ParseDuration(header); err != nil {
			r.cache.Write(key, bytes.NewReader(b), maxAge)
		} else {
			return err
		}
	}

	return nil
}

func (r *roundTripper) setProxyHeaders(resp *http.Response) {
	via := fmt.Sprintf("%s %s", resp.Proto, r.serverId)

	if v := resp.Header.Get("Via"); v != "" {
		resp.Header.Set("Via", v+", "+via)
	} else {
		resp.Header.Set("Via", via)
	}

	// squid sends these
	resp.Header.Del(CacheHeader)
	resp.Header.Del(CacheLookupHeader)
}

func copyResponse(resp *http.Response) (*http.Response, error) {
	resp2 := *resp // shallow copy is ok

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	resp.Body.Close()
	resp.Body = ioutil.NopCloser(bytes.NewReader(body))
	resp2.Body = ioutil.NopCloser(bytes.NewReader(body))

	return &resp2, nil
}

func (r *roundTripper) cacheHit(resp *http.Response) (*http.Response, error) {
	// set an Age header
	if t, err := http.ParseTime(resp.Header.Get("Date")); err == nil {
		diff := time.Now().Sub(t)
		resp.Header.Set("Age", fmt.Sprintf("%.f", diff.Seconds()))
	}

	r.setProxyHeaders(resp)
	resp.Header.Set(CacheHeader, "HIT from "+r.serverId)
	logResponse(resp)
	return resp, nil
}

func (r *roundTripper) cacheMiss(resp *http.Response) (*http.Response, error) {
	r.setProxyHeaders(resp)
	resp.Header.Set(CacheHeader, "MISS from "+r.serverId)
	logResponse(resp)
	return resp, nil
}

func (r *roundTripper) cacheSkip(resp *http.Response) (*http.Response, error) {
	r.setProxyHeaders(resp)
	resp.Header.Set(CacheHeader, "SKIP from "+r.serverId)
	logResponse(resp)
	return resp, nil
}

func isRequestCacheable(req *http.Request) bool {
	if req.Header.Get(MaxAgeHeader) == "" {
		return false
	}

	return req.Method == "GET" || req.Method == "HEAD"
}

func isResponseCacheable(resp *http.Response) bool {
	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound
}

// cacheKey returns an MD5 cache key for a request
func cacheKey(req *http.Request) string {
	url := req.URL.String()

	// canonical url is set upstream pre-rewrite
	if h := req.Header.Get(CanonicalUrlHeader); h != "" {
		url = h
	}

	return fmt.Sprintf("%x", md5.Sum([]byte(url)))
}

func logResponse(resp *http.Response) {
	status := resp.Header.Get(CacheHeader)

	if strings.HasPrefix(status, "HIT") {
		status = "\x1b[32;1mHIT\x1b[0m"
	} else if strings.HasPrefix(status, "MISS") {
		status = "\x1b[31;1mMISS\x1b[0m"
	} else {
		status = "\x1b[33;1mSKIP\x1b[0m"
	}

	log.Printf(
		"%s \"%s %s %s\" (%s) %d %s",
		resp.Request.RemoteAddr,
		resp.Request.Method,
		resp.Request.URL,
		resp.Request.Proto,
		resp.Status,
		resp.ContentLength,
		status,
	)
}
