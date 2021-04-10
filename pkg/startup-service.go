package pkg

import (
	"context"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/background"
	"github.com/wufe/polo/pkg/http/rest"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/services"
	"github.com/wufe/polo/pkg/storage"
	"github.com/wufe/polo/pkg/utils"
)

type Startup struct {
	environment   utils.Environment
	configuration *models.RootConfiguration
	applications  []*models.Application
	handler       *rest.Handler
	static        *services.StaticService
	appStorage    *storage.Application
	sesStorage    *storage.Session
	mediator      *background.Mediator
}

func NewStartup(
	environment utils.Environment,
	configuration *models.RootConfiguration,
	applications []*models.Application,
	handler *rest.Handler,
	static *services.StaticService,
	appStorage *storage.Application,
	sesStorage *storage.Session,
	mediator *background.Mediator) *Startup {
	return &Startup{
		environment:   environment,
		configuration: configuration,
		applications:  applications,
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
	s.watchApplications(context.Background())
	s.loadSessions()
	s.startSessions()
	s.static.LoadSessionHelper()
	s.startServer()
}

func (s *Startup) loadApplications() {
	for _, application := range s.applications {
		go func(a *models.Application) {
			err := s.mediator.ApplicationInit.Enqueue(a)
			if err != nil {
				log.Errorf("Error while loading application: %s", err.Error())
			}
		}(application)
	}
}

func (s *Startup) storeApplications() {
	for _, application := range s.applications {
		s.appStorage.Add(application)
	}
}

func (s *Startup) watchApplications(ctx context.Context) {
	for _, application := range s.applications {
		var filename string
		application.WithRLock(func(a *models.Application) {
			filename = a.Filename
		})
		conf := application.GetConfiguration()
		log.Infof("Watching file %s for app %s", filename, conf.Name)
		go func(filename string, application *models.Application, conf models.ApplicationConfiguration) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					time.Sleep(2 * time.Second)
					rootConfig, err := storage.UnmarshalConfiguration(filename, s.environment)
					if err != nil {
						continue
					}
					if rootConfig.ApplicationConfigurations != nil {
						for _, c := range rootConfig.ApplicationConfigurations {
							if c.Name == conf.Name {
								newConf := *c
								if !models.ConfigurationAreEqual(conf, newConf) {
									log.Infof(fmt.Sprintf("[APP:%s] Configuration changed", newConf.Name))
									application.SetConfiguration(newConf)
									conf = newConf
									sessions := s.sesStorage.GetByApplicationName(conf.Name)
									for _, session := range sessions {
										session.InitializeConfiguration()
									}
								}
							}
						}
					}
				}
			}
		}(filename, application, conf)
	}
}

func (s *Startup) loadSessions() {
	s.sesStorage.LoadSessions(s.appStorage, s.environment)
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
