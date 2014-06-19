package cache

import (
	"io"
	"os"
	"path/filepath"

	"github.com/peterbourgon/diskv"
)

type Cache interface {
	WriteStream(key string, r io.Reader, sync bool) error
	ReadStream(key string) (io.ReadCloser, error)
	Has(key string) bool
}

func NewFileCache(baseDir string) (Cache, error) {

	// file transform for disk cache
	t := func(s string) []string {
		fullPath := filepath.Join(baseDir, s)
		os.MkdirAll(filepath.Dir(fullPath), 0700)
		return []string{}
	}

	// basic disk cache
	d := diskv.New(diskv.Options{
		BasePath:     baseDir,
		Transform:    t,
		CacheSizeMax: 10 << 20,
	})

	return d, nil
}
