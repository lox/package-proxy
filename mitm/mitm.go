package mitm

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"

	gomitm "github.com/getlantern/go-mitm/mitm"
)

// default set of TLS cipher suites used
func defaultTlsConfig() *tls.Config {
	return &tls.Config{
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
			tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
			tls.TLS_RSA_WITH_RC4_128_SHA,
			tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
		PreferServerCipherSuites: true,
	}
}

// InterceptTlsHandler intercepts CONNECT requests and transparently decrypts them
// using a generated certificate
func InterceptTlsHandler(handler http.Handler, pkFile, certFile string) (*mitmHandler, error) {
	cryptoConfig := &gomitm.CryptoConfig{
		PKFile:          pkFile,
		CertFile:        certFile,
		ServerTLSConfig: defaultTlsConfig(),
	}

	wrapped, err := gomitm.Wrap(handler, cryptoConfig)
	if err != nil {
		return nil, err
	}

	return &mitmHandler{
		handler: handler, wrapped: wrapped, hosts: []string{},
	}, nil
}

type mitmHandler struct {
	handler, wrapped http.Handler
	hosts            []string
}

func (h *mitmHandler) AddHost(host string) {
	h.hosts = append(h.hosts, host)
}

func (h *mitmHandler) match(host string) bool {
	for _, h := range h.hosts {
		if h == host {
			return true
		}
	}

	return false
}

func (h *mitmHandler) connectProxy(rw http.ResponseWriter, req *http.Request) {
	remote, err := net.Dial("tcp", req.URL.Host)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	hj, ok := rw.(http.Hijacker)
	if !ok {
		log.Println("connection doesn't support hijacking")
		return
	}
	local, _, err := hj.Hijack()
	if err != nil {
		log.Println(err)
		return
	}

	local.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))

	go io.Copy(local, remote)
	go func() {
		io.Copy(remote, local)
		local.Close()
		remote.Close()
	}()
}

func (h *mitmHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.wrapped.ServeHTTP(rw, req)
	// if req.Method == "CONNECT" {
	// 	if h.match(req.URL.Host) {
	// 		log.Printf("intercepting CONNECT to %s", req.URL.Host)
	// 		h.wrapped.ServeHTTP(rw, req)
	// 	} else {
	// 		log.Printf("%s \"CONNECT %s %s\"", req.RemoteAddr, req.URL.Host, req.Proto)
	// 		h.connectProxy(rw, req)
	// 	}
	// } else {
	// 	h.handler.ServeHTTP(rw, req)
	// }
}
