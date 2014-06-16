package cache

import (
	"fmt"
	"regexp"
	"time"
)

type RefreshPattern struct {
	Pattern  string
	Duration string
}

type ParsedRefreshPattern struct {
	*regexp.Regexp
	pattern  string
	Duration time.Duration
}

func ParseRefreshPatterns(source []RefreshPattern) ([]ParsedRefreshPattern, error) {
	patterns := make([]ParsedRefreshPattern, len(source))

	for idx, p := range source {
		r, err := regexp.Compile(p.Pattern)
		if err != nil {
			return nil, err
		}
		d, err := time.ParseDuration(p.Duration)
		if err != nil {
			return nil, err
		}
		patterns[idx] = ParsedRefreshPattern{r, p.Pattern, d}
	}

	return patterns, nil
}

func (p *ParsedRefreshPattern) String() string {
	return fmt.Sprintf("%s %s", p.pattern, p.Duration.String())
}
