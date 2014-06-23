package cache

import (
	"io"
	"log"
	"os"
	f "path/filepath"
	"time"

	"github.com/peterbourgon/diskv"
)

const (
	defaultPrefix  = "default"
	expireInterval = time.Second * 5
	jsonFile       = "records.json"
	tmpBase        = "package-proxy"
)

// NewDiskCache creates a new disk-backed Cache in baseDir, if
// baseDir is an empty string it defaults to a system TempDir
func NewDiskCache(baseDir string, size uint64) (Cache, error) {
	if baseDir == "" {
		baseDir = f.Join(os.TempDir(), tmpBase)
	}

	d := diskv.New(diskv.Options{
		BasePath:     baseDir,
		CacheSizeMax: size,
		Transform: func(key string) []string {
			return []string{defaultPrefix, key[0:3]}
		},
	})

	expirer, err := LoadExpirer(f.Join(baseDir, jsonFile), func(key string) {
		log.Printf("expiring %s", key)
		d.Erase(key)
	})
	if err != nil {
		return nil, err
	}

	// kick off expiration
	go expirer.Tick(expireInterval)

	return &diskCache{
		diskv:   d,
		expirer: expirer,
	}, nil
}

type diskCache struct {
	diskv   *diskv.Diskv
	expirer *Expirer
}

func (c *diskCache) Write(key string, r io.Reader, maxAge time.Duration) error {
	c.expirer.SetLastUpdated(key, time.Now())
	c.expirer.SetMaxAge(key, maxAge)
	return c.diskv.WriteStream(key, r, true)
}

func (c *diskCache) Read(key string) (io.ReadCloser, error) {
	return c.diskv.ReadStream(key)
}

func (c *diskCache) Has(key string) bool {
	return c.diskv.Has(key)
}
