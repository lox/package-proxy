package cache

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
)

func RequestCacheKey(req *http.Request) string {
	return fmt.Sprintf("%s%s", req.URL.Host, req.URL.Path)
}

func CachedResponse(c Cache, req *http.Request) (resp *http.Response, err error) {
	cachedVal, ok := c.Get(RequestCacheKey(req))
	if !ok {
		return
	}

	b := bytes.NewBuffer(cachedVal)
	return http.ReadResponse(bufio.NewReader(b), req)
}
