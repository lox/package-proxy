package cache

import (
	"bytes"
	"io"
	"io/ioutil"
	"time"
)

func NewMapCache() *mapCache {
	return &mapCache{map[string][]byte{}}
}

type mapCache struct {
	Map map[string][]byte
}

func (m *mapCache) Write(key string, r io.Reader, maxAge time.Duration) error {
	buffer := bytes.Buffer{}
	_, err := buffer.ReadFrom(r)
	if err != nil {
		return err
	}

	m.Map[key] = buffer.Bytes()
	return nil
}

func (m *mapCache) Read(key string) (io.ReadCloser, error) {
	var buffer *bytes.Buffer

	if b, ok := m.Map[key]; ok {
		buffer = bytes.NewBuffer(b)
	} else {
		buffer = &bytes.Buffer{}
	}

	return ioutil.NopCloser(buffer), nil
}

func (m *mapCache) Has(key string) bool {
	_, ok := m.Map[key]
	return ok
}
