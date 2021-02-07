package net

import (
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/models"
	"github.com/wufe/polo/services"
	"github.com/wufe/polo/utils"
)

const (
// ServerRouteRoot                          ServerRoute = "/_polo_"
// ServerRouteDashboard                     ServerRoute = "/_polo_/"
// ServerRouteSessionStatus                 ServerRoute = "/_polo_/session/%s/" // uuid
// ServerRouteStatic                        ServerRoute = "/_polo_/public/*filepath"
// ServerRouteAPIApplications               ServerRoute = "/_polo_/api/application/"
// ServerRouteAPISession                    ServerRoute = "/_polo_/api/session/"
// ServerRouteAPISessionByUUID              ServerRoute = "/_polo_/api/session/:uuid"
// ServerRouteAPISessionAgeByUUID           ServerRoute = "/_polo_/api/session/:uuid/age"
// ServerRouteAPITrackSessionByUUID         ServerRoute = "/_polo_/api/session/:uuid/track"
// ServerRouteAPISessionLogsAndStatusByUUID ServerRoute = "/_polo_/api/session/:uuid/logs/:last_log"

// StaticFolderPath string = "/_polo_/public"

// StaticManagerFile       string = "/manager.html"
// StaticSessionHelperFile string = "/session-helper.html"
)

type HTTPServer struct {
	SessionHandler       *services.SessionHandler
	Configuration        *models.RootConfiguration
	sessionHelperContent string
	fileSystem           *http.FileSystem
	isDev                bool
	devServerURL         string
}

type ServerRoute string

func NewHTTPServer(port string, sessionHandler *services.SessionHandler, configuration *models.RootConfiguration, wg *sync.WaitGroup) *HTTPServer {

	server := &HTTPServer{
		SessionHandler:       sessionHandler,
		Configuration:        configuration,
		sessionHelperContent: "",
		isDev:                utils.IsDev(),
		devServerURL:         utils.DevServerURL(),
	}

	// server.initStaticFileSystem()

	// Session helper content retrieval

	wg.Add(1)
	go func() {
		router := httprouter.New()

		// router.GET(string(ServerRouteDashboard), server.getDashboard)
		// router.GET(strings.ReplaceAll(string(ServerRouteSessionStatus), "%s", ":uuid"), server.getSessionStatus)
		// router.GET(string(ServerRouteAPIApplications), server.getApplicationsAPI)
		// router.POST(string(ServerRouteAPISession), server.postSessionAPI)
		// router.GET(string(ServerRouteAPISession), server.getAllSessionsAPI)
		// router.GET(string(ServerRouteAPISessionByUUID), server.getSessionByUUIDAPI)
		// router.DELETE(string(ServerRouteAPISessionByUUID), server.deleteSessionByUUIDAPI)
		// router.GET(string(ServerRouteAPISessionAgeByUUID), server.getSessionAgeByUUIDAPI)
		// router.POST(string(ServerRouteAPITrackSessionByUUID), server.postTrackSessionByUUIDAPI)
		// router.DELETE(string(ServerRouteAPITrackSessionByUUID), server.postUntrackSessionAPI)
		// router.GET(string(ServerRouteAPISessionLogsAndStatusByUUID), server.getSessionLogsAndStatusByUUIDAPI)

		// server.serveStatic(router)

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
