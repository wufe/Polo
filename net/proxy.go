package net

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/models"
)

const (
	SESSION_COOKIE_NAME string = "PoloSession"
)

func (server *HTTPServer) getReverseProxyHandlerFunc() http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		isDev := true
		if isDev && (strings.HasPrefix(req.URL.Path, string(ServerRouteRoot)) ||
			strings.HasPrefix(req.URL.Path, "/sockjs-node")) {
			server.serveReverseProxy("http://localhost:9000/", res, req) // Webpack dev server
		} else {
			session := server.detectSession(req)
			if session == nil {
				server.temporaryRedirect(res, string(ServerRouteDashboard))
			} else {
				switch session.Status {
				case models.SessionStatusStarted:
					server.serveReverseProxy(session.Service.Target, res, req)
					break
				case models.SessionStatusStarting:
				default:
					server.trackSession(res, session)
					server.temporaryRedirect(res, fmt.Sprintf(string(ServerRouteSessionStatus), session.UUID))
					break
				}
			}
		}
	})
}

func (server *HTTPServer) detectSession(req *http.Request) *models.Session {
	cookie, err := req.Cookie(SESSION_COOKIE_NAME)
	if err == http.ErrNoCookie {
		return nil
	}
	sessionUUID := cookie.Value
	return server.SessionHandler.GetSessionByUUID(sessionUUID)
}

func (server *HTTPServer) trackSession(res http.ResponseWriter, session *models.Session) {
	res.Header().Add("Set-Cookie", SESSION_COOKIE_NAME+"="+session.UUID)
}

func (server *HTTPServer) temporaryRedirect(res http.ResponseWriter, location string) {
	res.Header().Add("Location", location)
	res.WriteHeader(307)
}

func (server *HTTPServer) serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {

	url, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ModifyResponse = func(res *http.Response) error {

		if res.Header.Get("Content-Type") == "text/html" {

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Printf("Error reading body: %v", err)
			}

			buffer := bytes.NewBufferString("<div style=\"position:fixed;bottom:0;right:0;padding:30px;z-index:9999;background:white;\">SESSION: TESTSESSION</div>")
			buffer.Write(body)

			res.Body = ioutil.NopCloser(buffer)
			res.Header["Content-Length"] = []string{fmt.Sprint(buffer.Len())}
		}

		return nil
	}

	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme

	req.Host = url.Host

	log.Println(req.URL.RequestURI())

	proxy.ServeHTTP(res, req)
}
