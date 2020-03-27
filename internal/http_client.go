package internal

import (
	"net/http"
	"net/http/httptest"
)

type Proxy struct {
	http.Handler
}

func NewProxyClient(base string) *http.Client {
	return &http.Client{
		Transport: &Proxy{
			Handler: http.FileServer(http.Dir(base)),
		},
	}
}

func (p *Proxy) RoundTrip(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	p.Handler.ServeHTTP(w, req)
	return w.Result(), nil
}
