package net

import (
	"net/http"
	"strings"
	"sync"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/models"
	"github.com/wufe/polo/services"
)

const (
	ServerRouteRoot          ServerRoute = "/_polo_"
	ServerRouteDashboard     ServerRoute = "/_polo_/"
	ServerRouteSessionStatus ServerRoute = "/_polo_/session/%s/" // uuid
	ServerRouteStatic        ServerRoute = "/_polo_/static/*filepath"
	ServerRouteAPIServices   ServerRoute = "/_polo_/api/service/"
	ServerRouteAPISession    ServerRoute = "/_polo_/api/session/"

	StaticFolder string = "./static"
)

type HTTPServer struct {
	SessionHandler *services.SessionHandler
	Configuration  *models.RootConfiguration
}

type ServerRoute string

func NewHTTPServer(port string, sessionHandler *services.SessionHandler, configuration *models.RootConfiguration, wg *sync.WaitGroup) *HTTPServer {
	server := &HTTPServer{
		SessionHandler: sessionHandler,
		Configuration:  configuration,
	}
	wg.Add(1)
	go func() {
		router := httprouter.New()

		router.GET(string(ServerRouteDashboard), server.getDashboard)
		router.GET(strings.ReplaceAll(string(ServerRouteSessionStatus), "%s", ":uuid"), server.getSessionStatus)
		router.GET(string(ServerRouteAPIServices), server.getServicesAPI)
		router.POST(string(ServerRouteAPISession), server.postSessionAPI)
		router.GET(string(ServerRouteAPISession), server.getAllSessionsAPI)

		// router.ServeFiles(string(ServerRouteStatic), http.Dir(StaticFolder))

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
