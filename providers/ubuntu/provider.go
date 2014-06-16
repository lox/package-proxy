package ubuntu

import (
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/lox/packageproxy/cache"
)

type ubuntuProvider struct {
	rewriteRegExp, hostRegExp *regexp.Regexp
	mirror                    *url.URL
	patterns                  []cache.ParsedRefreshPattern
}

var defaultMaxAge = time.Hour * 48

var refreshPatterns = []cache.RefreshPattern{
	cache.RefreshPattern{`deb$`, "72h"},
	cache.RefreshPattern{`udeb$`, "72h"},
	cache.RefreshPattern{`tar.gz$`, "72h"},
	cache.RefreshPattern{`DiffIndex$`, "48h"},
	cache.RefreshPattern{`PackagesIndex$`, "48h"},
	cache.RefreshPattern{`Packages\.(bz2|gz|lzma)$ `, "48h"},
	cache.RefreshPattern{`SourcesIndex$`, "48h"},
	cache.RefreshPattern{`Sources\.(bz2|gz|lzma)$`, "48h"},
	cache.RefreshPattern{`Release$`, "48h"},
	cache.RefreshPattern{`Translation-(en|fr)\.(gz|bz2|bzip2|lzma)$`, "48h"},
	cache.RefreshPattern{`Sources\.lzma$`, "48h"},
	cache.RefreshPattern{`Sources\.lzma$`, "48h"},
}

var hosts = []string{
	`ppa.launchpad.net`,
	`security.ubuntu.com`,
	`archive.ubuntu.com`,
}

func NewProvider() *ubuntuProvider {
	u := &ubuntuProvider{}

	patterns, err := cache.ParseRefreshPatterns(refreshPatterns)
	if err != nil {
		log.Fatal(err)
	}

	u.hostRegExp = regexp.MustCompile(
		strings.Replace(strings.Join(hosts, "|"), ".", "\\.", -1),
	)

	u.rewriteRegExp = regexp.MustCompile(
		`https?://(security|archive).ubuntu.com/ubuntu/(.+)$`,
	)

	u.patterns = patterns

	mirrors, err := GetGeoMirrors()
	if err != nil {
		log.Fatal(err)
	}

	random := rand.Intn(len(mirrors.URLs))
	u.mirror, _ = url.Parse(mirrors.URLs[random])

	log.Printf("using ubuntu mirror %s whilst benchmarking", u.mirror.String())

	// benchmark in the background to make sure we have the fastest
	go func(u *ubuntuProvider) {
		mirror, err := mirrors.Fastest()
		if err != nil {
			log.Println(err)
		}

		if mirrorUrl, err := url.Parse(mirror); err == nil {
			log.Printf("benchmark complete, using ubuntu mirror %s", mirror)
			u.mirror = mirrorUrl
		}
	}(u)

	return u
}

func (p *ubuntuProvider) Match(r *http.Request) bool {
	return p.hostRegExp.MatchString(r.URL.String())
}

func (p *ubuntuProvider) RewriteUrl(u *url.URL) {
	if p.rewriteRegExp.MatchString(u.String()) {
		m := p.rewriteRegExp.FindAllStringSubmatch(u.String(), -1)
		u.Host = p.mirror.Host
		u.Path = p.mirror.Path + m[0][2]
	}
}

func (p *ubuntuProvider) CacheKey(r *http.Request) string {
	if p.rewriteRegExp.MatchString(r.URL.String()) {
		m := p.rewriteRegExp.FindAllStringSubmatch(r.URL.String(), -1)
		return "ubuntu/" + p.mirror.Path + m[0][2]
	} else {
		return "ubuntu/ppa/" + r.URL.Host + r.URL.Path
	}
}

func (p *ubuntuProvider) CacheMaxAge(r *http.Request) time.Duration {
	for _, pattern := range p.patterns {
		if pattern.MatchString(r.URL.String()) {
			return pattern.Duration
		}
	}

	return defaultMaxAge
}
