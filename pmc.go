package main

import (
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"

	"github.com/99designs/pmc/cache"
	"github.com/99designs/pmc/providers/ubuntu"
)

var ubuntuRegex *regexp.Regexp

func init() {
	ubuntuRegex = regexp.MustCompile(`/(security|archive)\.ubuntu\.com/ubuntu/(.+)$`)
}

type Pmc struct {
	Proxy        *httputil.ReverseProxy
	Cache        cache.Cache
	UbuntuMirror *url.URL
}

func (p *Pmc) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	p.Proxy.ServeHTTP(rw, req)
}

func NewPmc() *Pmc {
	c := cache.NewMemoryCache()
	proxy := &httputil.ReverseProxy{
		Transport: &cache.RoundTripper{http.DefaultTransport, c},
	}

	pmc := &Pmc{Proxy: proxy, Cache: c}
	pmc.Proxy.Director = func(r *http.Request) {

		// rewrite ubuntu urls to a mirror
		if ubuntuRegex.MatchString(r.URL.String()) {
			if pmc.UbuntuMirror != nil {
				m := ubuntuRegex.FindAllStringSubmatch(r.URL.String(), -1)
				r.URL.Host = pmc.UbuntuMirror.Host
				r.URL.Path = pmc.UbuntuMirror.Path + m[0][2]
			}
		}
	}

	go func() {
		log.Printf("benchmarking ubuntu mirrors")
		mirrors, err := ubuntu.GetGeoMirrors()
		if err != nil {
			log.Println(err)
		}

		random := rand.Intn(len(mirrors.URLs))
		pmc.UbuntuMirror, _ = url.Parse(mirrors.URLs[random])
		log.Printf("using mirror %s whilst benchmarking", pmc.UbuntuMirror.String())

		u, err := mirrors.Fastest()
		if err != nil {
			log.Println(err)
		}

		log.Printf("benchmark done, picking %s", u)
		pmc.UbuntuMirror, _ = url.Parse(u)

	}()

	return pmc
}

func main() {
	log.Println("listening on :3142")
	log.Fatal(http.ListenAndServe(":3142", NewPmc()))
}
