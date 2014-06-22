package api

import (
	"log"
	"net/http"
)

func NewApiHandler(delegate http.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/package-proxy.pem", func(rw http.ResponseWriter, r *http.Request) {
		log.Printf("%s requested package-proxy.pem", r.RemoteAddr)
		http.ServeFile(rw, r, "certs/public.pem")

	})

	return &apiHandler{delegate, mux}
}

type apiHandler struct {
	delegate, mux http.Handler
}

func (h *apiHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "CONNECT" {
		h.delegate.ServeHTTP(rw, req)
	} else {
		h.mux.ServeHTTP(rw, req)
	}
}
