package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/utils"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Handler struct {
	isDev         bool
	devServerURL  string
	log           logging.Logger
	configuration *models.RootConfiguration
}

type Builder func(url *url.URL) *httputil.ReverseProxy

var DefaultReverseProxy Builder = func(url *url.URL) *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(url)
}

func NewHandler(environment utils.Environment, logger logging.Logger, configuration *models.RootConfiguration) *Handler {
	s := &Handler{
		isDev:         environment.IsDev(),
		devServerURL:  environment.DevServerURL(),
		log:           logger,
		configuration: configuration,
	}
	return s
}

func (s *Handler) ServeDevServer(w http.ResponseWriter, r *http.Request) {
	s.ServeDefaultReverseProxy(s.devServerURL, w, r)
}

func (s *Handler) ServeDefaultReverseProxy(target string, w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(target)
	if err != nil {
		s.log.Errorf("Error creating target url: %s", err.Error())
	}
	proxy := DefaultReverseProxy(u)
	proxy.ModifyResponse = func(response *http.Response) error {
		if strings.Contains(r.URL.String(), "manager.html") {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return fmt.Errorf("error reading the response body: %w", err)
			}
			defer response.Body.Close()

			serializedConfiguration, err := json.Marshal(s.configuration.GetManagerConfiguration())
			if err != nil {
				return fmt.Errorf("error serializing configuration: %w", err)
			}
			body = []byte(strings.ReplaceAll(string(body), "{}", string(serializedConfiguration)))

			contentLength := len(body)
			response.Body = io.NopCloser(bytes.NewReader(body))
			response.Header["Content-Length"] = []string{fmt.Sprint(contentLength)}
		}
		return nil
	}

	s.Serve(proxy)(w, r)
}

func (s *Handler) Serve(proxy *httputil.ReverseProxy) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}
