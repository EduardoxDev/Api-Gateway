package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Proxy struct {
	target *url.URL
	proxy  *httputil.ReverseProxy
}

func NewProxy(target string) (*Proxy, error) {
	url, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	// Customize the proxy director to handle headers correctly
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = url.Host
	}

	return &Proxy{
		target: url,
		proxy:  proxy,
	}, nil
}

func (p *Proxy) TargetHost() string {
	return p.target.Host
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request, pathPrefix string) {
	// Strip prefix before forwarding
	r.URL.Path = strings.TrimPrefix(r.URL.Path, pathPrefix)
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}
	p.proxy.ServeHTTP(w, r)
}
