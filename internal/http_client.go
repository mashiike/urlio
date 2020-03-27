package internal

import (
	"net/http"
	"net/http/httptest"
)

type Proxy struct {
	http.Handler
	checkHost bool
}

func NewProxyClient(base string, checkHost bool) *http.Client {
	return &http.Client{
		Transport: &Proxy{
			Handler:   http.FileServer(http.Dir(base)),
			checkHost: checkHost,
		},
	}
}

func (p *Proxy) RoundTrip(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	if p.checkHost {
		req.URL.Path = req.URL.Host + "/" + req.URL.Path
	}
	p.Handler.ServeHTTP(w, req)
	return w.Result(), nil
}
