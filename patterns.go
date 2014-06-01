package main

import (
	"regexp"
	"time"
)

// derived from https://tails.boum.org/contribute/build/squid-deb-proxy/squid-deb-proxy.conf

type RefreshPattern struct {
	Pattern  string
	Duration string
}

var refreshPatterns = []RefreshPattern{
	RefreshPattern{`deb$`, "2 days"},
	RefreshPattern{`udeb$`, "2 days"},
	RefreshPattern{`tar.gz$`, "2 days"},
	RefreshPattern{`DiffIndex$`, "2 days"},
	RefreshPattern{`PackagesIndex$`, "2 days"},
	RefreshPattern{`Packages\.(bz2|gz|lzma)$ `, "2 days"},
	RefreshPattern{`SourcesIndex$`, "2 days"},
	RefreshPattern{`Sources\.(bz2|gz|lzma)$`, "2 days"},
	RefreshPattern{`Release$`, "2 days"},
	RefreshPattern{`Translation-(en|fr)\.(gz|bz2|bzip2|lzma)$`, "2 days"},
	RefreshPattern{`Sources\.lzma$`, "2 days"},
	RefreshPattern{`Sources\.lzma$`, "2 days"},
}

type ParsedRefreshPattern struct {
	Pattern  regexp.Regexp
	Duration time.Duration
}

func ParseRefreshPatterns() ([]ParsedRefreshPattern, error) {
	patterns := [len(refreshPatterns)]ParsedRefreshPattern{}

	for idx, p := range refreshPatterns {
		r, err := regexp.Compile(p.Pattern)
		if err != nil {
			return nil, err
		}
		d, err := time.ParseDuration(p.Duration)
		if err != nil {
			return nil, err
		}
		patterns[idx] = ParsedRefreshPattern{r, d}
	}

	return patterns
}
