package cache

import (
	"io"

	"time"
)

type Cache interface {
	Write(key string, r io.Reader, maxAge time.Duration) error
	Read(key string) (io.ReadCloser, error)
	Has(key string) bool
}
