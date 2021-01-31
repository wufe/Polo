package net

import (
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/models"
	"github.com/wufe/polo/services"
	"github.com/wufe/polo/utils"
)

const (
	ServerRouteRoot                  ServerRoute = "/_polo_"
	ServerRouteDashboard             ServerRoute = "/_polo_/"
	ServerRouteSessionStatus         ServerRoute = "/_polo_/session/%s/" // uuid
	ServerRouteStatic                ServerRoute = "/_polo_/static/*filepath"
	ServerRouteAPIServices           ServerRoute = "/_polo_/api/service/"
	ServerRouteAPISession            ServerRoute = "/_polo_/api/session/"
	ServerRouteAPISessionByUUID      ServerRoute = "/_polo_/api/session/:uuid"
	ServerRouteAPISessionAgeByUUID   ServerRoute = "/_polo_/api/session/:uuid/age"
	ServerRouteAPITrackSessionByUUID ServerRoute = "/_polo_/api/session/:uuid/track"
	ServerRouteAPIUntrackSession     ServerRoute = "/_polo_/api/session/track"

	StaticFolder string = "./static"
)

type HTTPServer struct {
	SessionHandler       *services.SessionHandler
	Configuration        *models.RootConfiguration
	sessionHelperContent string
}

type ServerRoute string

func NewHTTPServer(port string, sessionHandler *services.SessionHandler, configuration *models.RootConfiguration, wg *sync.WaitGroup) *HTTPServer {

	server := &HTTPServer{
		SessionHandler:       sessionHandler,
		Configuration:        configuration,
		sessionHelperContent: "",
	}

	// Session helper content retrieval
	if utils.IsDev() {
		// If in dev mode, the content is available via webpack dev server
		go func() {
			for {
				resp, err := http.Get("http://localhost:9000/_polo_/static/session-helper.html")
				if err != nil {
					log.Errorf("Error while getting session helper: %s", err.Error())
				} else {
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Errorf("Error while reading session helper response: %s", err.Error())
					} else {
						server.sessionHelperContent = string(body)
					}
					resp.Body.Close()
				}

				time.Sleep(30 * time.Second)
			}
		}()
	} else {
		file, err := http.Dir(StaticFolder).Open("session-helper.html")
		if err != nil {
			log.Errorf("Error while getting session helper: %s", err.Error())
		} else {
			defer file.Close()
			content, err := ioutil.ReadAll(file)
			if err != nil {
				log.Errorf("Error while reading session helper content: %s", err.Error())
			} else {
				server.sessionHelperContent = string(content)
			}
		}
	}

	wg.Add(1)
	go func() {
		router := httprouter.New()

		router.GET(string(ServerRouteDashboard), server.getDashboard)
		router.GET(strings.ReplaceAll(string(ServerRouteSessionStatus), "%s", ":uuid"), server.getSessionStatus)
		router.GET(string(ServerRouteAPIServices), server.getServicesAPI)
		router.POST(string(ServerRouteAPISession), server.postSessionAPI)
		router.GET(string(ServerRouteAPISession), server.getAllSessionsAPI)
		router.GET(string(ServerRouteAPISessionByUUID), server.getSessionByUUIDAPI)
		router.GET(string(ServerRouteAPISessionAgeByUUID), server.getSessionAgeByUUIDAPI)
		router.POST(string(ServerRouteAPITrackSessionByUUID), server.postTrackSessionByUUIDAPI)
		router.DELETE(string(ServerRouteAPIUntrackSession), server.postUntrackSessionAPI)

		server.serveStatic(router)

		router.NotFound = server.getReverseProxyHandlerFunc()

		server := &http.Server{
			Addr:    ":" + port,
			Handler: router,
		}

		log.Infof("Starting reverse proxy server on port %s..", port)

		if port == "443" {
			if err := server.ListenAndServeTLS(configuration.Global.TLSCertFile, configuration.Global.TLSKeyFile); err != http.ErrServerClosed {
				panic(err)
			}
		} else {
			if err := server.ListenAndServe(); err != http.ErrServerClosed {
				panic(err)
			}
		}

		wg.Done()
	}()

	return server
}
