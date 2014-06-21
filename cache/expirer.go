package cache

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

const (
	defaultMaxAge = time.Duration(0)
)

type ExpireFunc func(key string)

type Expirer struct {
	timer      <-chan time.Time
	records    map[string]keyRecord
	mutex      sync.RWMutex
	expireFunc ExpireFunc
	jsonFile   string
	dirty      bool
}

type keyRecord struct {
	MaxAge      time.Duration
	TimeUpdated time.Time
}

func (r *keyRecord) TimeToLive() time.Duration {
	return r.TimeUpdated.Add(r.MaxAge).Sub(time.Now())
}

func NewExpirer(e ExpireFunc) *Expirer {
	return &Expirer{
		records:    map[string]keyRecord{},
		expireFunc: e,
	}
}

func LoadExpirer(jsonFile string, e ExpireFunc) (*Expirer, error) {
	expirer := NewExpirer(e)
	expirer.jsonFile = jsonFile

	_, err := os.Stat(jsonFile)
	if err == nil {
		jsonBlob, err := ioutil.ReadFile(jsonFile)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(jsonBlob, &expirer.records)
		if err != nil {
			return nil, err
		}

		log.Printf("loaded %d records from %s", len(expirer.records), jsonFile)
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	return expirer, nil
}

func (e *Expirer) Save() error {
	log.Printf("saving %d records to %s", len(e.records), e.jsonFile)

	jsonBlob, err := json.Marshal(e.records)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(e.jsonFile, jsonBlob, 0777)
	if err != nil {
		return err
	}

	e.dirty = false
	return nil
}

func (e *Expirer) SetLastUpdated(key string, t time.Time) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if _, ok := e.records[key]; ok {
		r := e.records[key]
		r.TimeUpdated = t
		e.records[key] = r
	} else {
		e.records[key] = keyRecord{
			TimeUpdated: t,
			MaxAge:      defaultMaxAge,
		}
	}

	e.dirty = true
}

func (e *Expirer) SetMaxAge(key string, d time.Duration) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if _, ok := e.records[key]; ok {
		r := e.records[key]
		r.MaxAge = d
		e.records[key] = r
	} else {
		e.records[key] = keyRecord{
			MaxAge: d,
		}
	}

	e.dirty = true
}

func (e *Expirer) Expire(t time.Time) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	for key, r := range e.records {
		ttl := r.TimeToLive()
		if ttl <= 0 {
			e.expireFunc(key)
			delete(e.records, key)
		}
	}
}

func (e *Expirer) Tick(d time.Duration) {
	e.timer = time.Tick(d)
	for now := range e.timer {
		e.Expire(now)
		if e.dirty {
			e.Save()
		}
	}
}
