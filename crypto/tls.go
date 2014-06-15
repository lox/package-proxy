package crypto

import (
	"crypto/tls"
	"net/http"

	"github.com/getlantern/go-mitm/mitm"
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

// UnwrapTlsHandler transparently unwraps TLS from an http.Handler via
// X.509 certificates signed with the provided private/public key pair
func UnwrapTlsHandler(handler http.Handler, pkFile, certFile string) (http.Handler, error) {
	cryptoConfig := &mitm.CryptoConfig{
		PKFile:          pkFile,
		CertFile:        certFile,
		ServerTLSConfig: defaultTlsConfig(),
	}

	return mitm.Wrap(handler, cryptoConfig)
}
