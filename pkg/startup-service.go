package pkg

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/background"
	"github.com/wufe/polo/pkg/http/rest"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/services"
	"github.com/wufe/polo/pkg/storage"
)

type Startup struct {
	isDev         bool
	configuration *models.RootConfiguration
	handler       *rest.Handler
	static        *services.StaticService
	appStorage    *storage.Application
	sesStorage    *storage.Session
	mediator      *background.Mediator
}

func NewStartup(
	isDev bool,
	configuration *models.RootConfiguration,
	handler *rest.Handler,
	static *services.StaticService,
	appStorage *storage.Application,
	sesStorage *storage.Session,
	mediator *background.Mediator) *Startup {
	return &Startup{
		isDev:         isDev,
		configuration: configuration,
		handler:       handler,
		static:        static,
		appStorage:    appStorage,
		sesStorage:    sesStorage,
		mediator:      mediator,
	}
}

func (s *Startup) Start() {
	s.loadApplications()
	s.storeApplications()
	s.loadSessions()
	s.startSessions()
	s.static.LoadSessionHelper()
	s.startServer()
}

func (s *Startup) loadApplications() {
	for _, application := range s.configuration.Applications {
		s.mediator.ApplicationInit.Enqueue(application)
	}
}

func (s *Startup) storeApplications() {
	for _, application := range s.configuration.Applications {
		s.appStorage.Add(application)
	}
}

func (s *Startup) loadSessions() {
	s.sesStorage.LoadSessions(s.appStorage)
}

func (s *Startup) startSessions() {
	for _, session := range s.sesStorage.GetAllAliveSessions() {
		s.mediator.HealthcheckSession.Enqueue(session)
	}
}

func (s *Startup) startServer() {
	port := fmt.Sprint(s.configuration.Global.Port)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: s.handler.Router,
	}

	log.Infof("Server started on port %s", port)
	if port == "443" {
		if err := server.ListenAndServeTLS(s.configuration.Global.TLSCertFile, s.configuration.Global.TLSKeyFile); err != http.ErrServerClosed {
			panic(err)
		}
	} else {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}
}
