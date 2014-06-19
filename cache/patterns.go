package cache

import (
	"regexp"
	"time"
)

// NewPattern creates a new cache pattern, panics on parse error
func NewPattern(pattern string, d time.Duration) *cachePattern {
	return &cachePattern{Regexp: regexp.MustCompile(pattern), Duration: d}
}

type cachePattern struct {
	*regexp.Regexp
	Duration time.Duration
}

type CachePatternSlice []*cachePattern

// MatchString tries to match a given string across all patterns
func (r CachePatternSlice) MatchString(subject string) (bool, *cachePattern) {
	for _, p := range r {
		if p.MatchString(subject) {
			return true, p
		}
	}

	return false, nil
}
