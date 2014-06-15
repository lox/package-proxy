package ubuntu

import (
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
)

func NewProvider() *ubuntuProvider {
	u := &ubuntuProvider{}
	u.regex = regexp.MustCompile(`/(security|archive)\.ubuntu\.com/ubuntu/(.+)$`)

	mirrors, err := GetGeoMirrors()
	if err != nil {
		log.Println(err)
	}

	random := rand.Intn(len(mirrors.URLs))
	u.mirrorHost, _ = url.Parse(mirrors.URLs[random])

	log.Printf("using ubuntu mirror %s whilst benchmarking", u.mirrorHost.String())

	// benchmark in the background to make sure we have the fastest
	go func(u *ubuntuProvider) {
		mirror, err := mirrors.Fastest()
		if err != nil {
			log.Println(err)
		}

		if mirrorUrl, err := url.Parse(mirror); err == nil {
			log.Printf("benchmark complete, using ubuntu mirror %s", mirror)
			u.mirrorHost = mirrorUrl
		}
	}(u)

	return u
}

type ubuntuProvider struct {
	regex      *regexp.Regexp
	mirrorHost *url.URL
	mirrorPath string
}

func (p *ubuntuProvider) Match(r *http.Request) bool {
	return p.regex.MatchString(r.URL.String())
}

func (p *ubuntuProvider) Rewrite(r *http.Request) {
	m := p.regex.FindAllStringSubmatch(r.URL.String(), -1)
	r.URL.Host = p.mirrorHost.Host
	r.URL.Path = p.mirrorPath + m[0][2]
}
