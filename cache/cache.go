package cache

import "io"

type Cache interface {
	WriteStream(key string, r io.Reader, sync bool) error
	ReadStream(key string) (io.ReadCloser, error)
	Has(key string) bool
	Erase(key string) error
}
