package net

import (
	"bytes"
	"encoding/json"
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
		if server.isDev && (strings.HasPrefix(req.URL.Path, string(ServerRouteRoot)) ||
			// Webpack dev server
			strings.HasPrefix(req.URL.Path, "/sockjs-node")) {
			server.serveReverseProxy(server.devServerURL, res, req, nil)
		} else {
			session := server.detectSession(req)
			if session == nil {
				server.temporaryRedirect(res, string(ServerRouteDashboard))
			} else {
				switch session.Status {
				case models.SessionStatusStarted:
					server.serveReverseProxy(session.Target, res, req, session)
					break
				case models.SessionStatusStarting:
					server.trackSession(res, session)
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

func (server *HTTPServer) detectSession(req *http.Request) *models.Session {
	cookie, err := req.Cookie(SESSION_COOKIE_NAME)
	if err == http.ErrNoCookie {
		return nil
	}
	sessionUUID := cookie.Value
	return server.SessionHandler.GetSessionByUUID(sessionUUID)
}

func (server *HTTPServer) trackSession(res http.ResponseWriter, session *models.Session) {

	cookie := http.Cookie{
		Name:     SESSION_COOKIE_NAME,
		Value:    session.UUID,
		Path:     "/",
		MaxAge:   60 * 60 * 24,
		HttpOnly: true,
	}

	http.SetCookie(res, &cookie)
}

func (server *HTTPServer) untrackSession(res http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     SESSION_COOKIE_NAME,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}

	http.SetCookie(res, &cookie)
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

			if res.Header.Get("Content-Type") == "text/html" {

				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					log.Printf("Error reading body: %v", err)
				}

				serializedSession, err := json.Marshal(session)
				if err != nil {
					serializedSession = []byte(`{}`)
				}

				sessionHelper := strings.ReplaceAll(server.sessionHelperContent, "%%currentSession%%", string(serializedSession))
				sessionHelper = strings.ReplaceAll(sessionHelper, `\\`, `\\\\`)

				buffer := bytes.NewBufferString(sessionHelper)

				// buffer := bytes.NewBufferString(fmt.Sprintf("<div style=\"position:fixed;bottom:0;right:0;padding:30px;z-index:9999;background:white;\">SESSION: %s</div>", session.UUID))
				buffer.Write(body)

				res.Body = ioutil.NopCloser(buffer)
				res.Header["Content-Length"] = []string{fmt.Sprint(buffer.Len())}
			}

			return nil
		}
	}

	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme

	req.Host = url.Host

	if session.Service.Host != "" {
		req.Header.Add("Host", session.Service.Host)
	}

	if session != nil {
		log.Printf("[PROXY -> SESSION:%s] %s", session.UUID, req.URL.RequestURI())
	} else {
		log.Printf("[PROXY -> _POLO_] %s", req.URL.RequestURI())
	}

	proxy.ServeHTTP(res, req)
}
