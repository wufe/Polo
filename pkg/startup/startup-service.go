package startup

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/background"
	"github.com/wufe/polo/pkg/http/rest"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/static"
	"github.com/wufe/polo/pkg/storage"
)

type Service struct {
	isDev         bool
	configuration *models.RootConfiguration
	handler       *rest.Handler
	static        *static.Service
	appStorage    *storage.Application
	mediator      *background.Mediator
}

func NewService(
	isDev bool,
	configuration *models.RootConfiguration,
	handler *rest.Handler,
	static *static.Service,
	appStorage *storage.Application,
	mediator *background.Mediator) *Service {
	return &Service{
		isDev:         isDev,
		configuration: configuration,
		handler:       handler,
		static:        static,
		appStorage:    appStorage,
		mediator:      mediator,
	}
}

func (s *Service) Start() {
	s.loadApplications()
	s.storeApplications()
	s.static.LoadSessionHelper()
	s.startServer()
}

func (s *Service) loadApplications() {
	for _, application := range s.configuration.Applications {
		s.mediator.ApplicationInit.Request(application)
	}
}

func (s *Service) storeApplications() {
	for _, application := range s.configuration.Applications {
		s.appStorage.Add(application)
	}
}

func (s *Service) startServer() {
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
