package net

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/models"
)

const ()

// http://localhost:8888/monito/d8cf5838-fa3a-4c63-b82e-a0c3fe46f402

func (server *HTTPServer) getReverseProxyHandlerFunc() http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if server.isDev && (strings.HasPrefix(req.URL.Path, string(ServerRouteRoot)) ||
			// Webpack dev server
			strings.HasPrefix(req.URL.Path, "/sockjs-node")) {
			server.serveReverseProxy(server.devServerURL, res, req, nil)
		} else {
			usingSmartURL := false
			session := server.detectSession(req)
			if session == nil {
				session = server.tryGetSessionByRequestURL(req)
				usingSmartURL = true
			}
			if session == nil {
				server.temporaryRedirect(res, string(ServerRouteDashboard))
			} else {
				switch session.Status {
				case models.SessionStatusStarted:
					server.trackSession(res, session)
					if usingSmartURL {
						server.temporaryRedirect(res, "/")
					} else {
						target := findReverseProxyTargetByRequestPath(req, session)
						fmt.Println("target", target, "path", req.URL.Path, "scheme", req.URL.Scheme)
						server.serveReverseProxy(target, res, req, session)
					}
					break
				case models.SessionStatusStarting:
					server.temporaryRedirect(res, fmt.Sprintf(string(ServerRouteSessionStatus), session.UUID))
					break
				default:
					server.untrackSession(res)
					server.temporaryRedirect(res, string(ServerRouteDashboard))
					break
				}
			}
		}
	})
}

func (server *HTTPServer) temporaryRedirect(res http.ResponseWriter, location string) {
	res.Header().Add("Location", location)
	res.WriteHeader(307)
}

func (server *HTTPServer) serveReverseProxy(target string, res http.ResponseWriter, req *http.Request, session *models.Session) {

	url, err := url.Parse(target)
	if err != nil {
		log.Errorf("Error creating target url: %s", err.Error())
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	if session != nil {
		server.SessionHandler.MarkSessionAsBeingRequested(session)
		proxy.ModifyResponse = func(res *http.Response) error {

			return nil
		}
	}

	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme

	req.Host = url.Host

	if session != nil {
		if session.Application.Host != "" {
			req.Header.Add("Host", session.Application.Host)
			req.Host = session.Application.Host
		}
		log.Printf("[PROXY -> SESSION:%s] %s", session.UUID, req.URL.RequestURI())
	} else {
		log.Printf("[PROXY -> _POLO_] %s", req.URL.RequestURI())
	}

	proxy.ServeHTTP(res, req)
}
