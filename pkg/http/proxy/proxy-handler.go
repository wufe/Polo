package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/utils"
)

type Handler struct {
	isDev        bool
	devServerURL string
}

type Builder func(url *url.URL) *httputil.ReverseProxy

var DefaultReverseProxy Builder = func(url *url.URL) *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(url)
}

func NewHandler(environment utils.Environment) *Handler {
	s := &Handler{
		isDev:        environment.IsDev(),
		devServerURL: environment.DevServerURL(),
	}
	return s
}

func (s *Handler) ServeDevServer(w http.ResponseWriter, r *http.Request) {
	s.ServeDefaultReverseProxy(s.devServerURL, w, r)
}

func (s *Handler) ServeDefaultReverseProxy(target string, w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(target)
	if err != nil {
		log.Errorf("Error creating target url: %s", err.Error())
	}
	proxy := DefaultReverseProxy(u)
	s.Serve(proxy)(w, r)
}

func (s *Handler) Serve(proxy *httputil.ReverseProxy) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}
