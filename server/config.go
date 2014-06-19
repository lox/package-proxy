package server

type Config struct {
	EnableTls bool
	CacheDir  string
	Rewriters []Rewriter
}
