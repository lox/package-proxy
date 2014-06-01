package cache

import "net/http"

type Cache interface {
	Get(key string) (resp []byte, ok bool)
	Set(key string, resp []byte)
	Delete(key string)
}

func cacheKey(req *http.Request) string {
	return req.URL.String()
}
