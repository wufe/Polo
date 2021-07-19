package pkg

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/wufe/polo/pkg/background"
	"github.com/wufe/polo/pkg/background/queues"
	"github.com/wufe/polo/pkg/http/rest"
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/services"
	"github.com/wufe/polo/pkg/storage"
	"go.uber.org/dig"
)

type Startup struct {
	configuration      *models.RootConfiguration
	applications       []*models.Application
	handler            *rest.Handler
	static             *services.StaticService
	appStorage         *storage.Application
	sesStorage         *storage.Session
	mediator           *background.Mediator
	applicationBuilder *models.ApplicationBuilder
	sessionBuilder     *models.SessionBuilder
	log                logging.Logger

	sessionBuildWorker       *background.SessionBuildWorker
	sessionStartWorker       *background.SessionStartWorker
	sessionCleanWorker       *background.SessionCleanWorker
	sessionFilesystemWorker  *background.SessionFilesystemWorker
	sessionDestroyWorker     *background.SessionDestroyWorker
	sessionHealthcheckWorker *background.SessionHealthcheckWorker
	applicationInitWorker    *background.ApplicationInitWorker
	applicationFetchWorker   *background.ApplicationFetchWorker
}

type StartupParams struct {
	dig.In
	Configuration      *models.RootConfiguration
	Applications       []*models.Application
	Handler            *rest.Handler
	Static             *services.StaticService
	AppStorage         *storage.Application
	SesStorage         *storage.Session
	Mediator           *background.Mediator
	ApplicationBuilder *models.ApplicationBuilder
	SessionBuilder     *models.SessionBuilder
	Logger             logging.Logger

	SessionBuildWorker       *background.SessionBuildWorker
	SessionStartWorker       *background.SessionStartWorker
	SessionCleanWorker       *background.SessionCleanWorker
	SessionFilesystemWorker  *background.SessionFilesystemWorker
	SessionDestroyWorker     *background.SessionDestroyWorker
	SessionHealthcheckWorker *background.SessionHealthcheckWorker
	ApplicationInitWorker    *background.ApplicationInitWorker
	ApplicationFetchWorker   *background.ApplicationFetchWorker
}

type StartupOptions struct {
	WatchApplications bool
	LoadSessionHelper bool
	StartServer       bool
}

func NewStartup(params StartupParams) *Startup {
	return &Startup{
		configuration:      params.Configuration,
		applications:       params.Applications,
		handler:            params.Handler,
		static:             params.Static,
		appStorage:         params.AppStorage,
		sesStorage:         params.SesStorage,
		mediator:           params.Mediator,
		applicationBuilder: params.ApplicationBuilder,
		sessionBuilder:     params.SessionBuilder,
		log:                params.Logger,

		sessionBuildWorker:       params.SessionBuildWorker,
		sessionStartWorker:       params.SessionStartWorker,
		sessionCleanWorker:       params.SessionCleanWorker,
		sessionFilesystemWorker:  params.SessionFilesystemWorker,
		sessionDestroyWorker:     params.SessionDestroyWorker,
		sessionHealthcheckWorker: params.SessionHealthcheckWorker,
		applicationInitWorker:    params.ApplicationInitWorker,
		applicationFetchWorker:   params.ApplicationFetchWorker,
	}
}

func (s *Startup) Start(options *StartupOptions) {
	if options == nil {
		options = &StartupOptions{
			WatchApplications: true,
			LoadSessionHelper: true,
			StartServer:       true,
		}
	}

	s.sessionBuildWorker.Start()
	s.sessionStartWorker.Start()
	s.sessionCleanWorker.Start()
	s.sessionFilesystemWorker.Start()
	s.sessionDestroyWorker.Start()
	s.sessionHealthcheckWorker.Start()
	s.applicationInitWorker.Start()
	s.applicationFetchWorker.Start()

	s.loadApplications()
	s.storeApplications()
	if options.WatchApplications {
		s.watchApplications(context.Background())
	}
	s.loadSessions()
	s.startSessions()
	if options.LoadSessionHelper {
		s.static.LoadSessionHelper()
	}
	if options.StartServer {
		s.startServer()
	}
}

func (s *Startup) loadApplications() {
	for _, application := range s.applications {
		go func(a *models.Application) {
			err := s.mediator.ApplicationInit.Enqueue(a)
			if err != nil {
				s.log.Errorf("Error while loading application: %s", err.Error())
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
		s.log.Infof("Watching file %s for app %s", filename, conf.Name)
		go func(filename string, application *models.Application, conf models.ApplicationConfiguration) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					time.Sleep(2 * time.Second)
					rootConfig, err := storage.UnmarshalConfiguration(filename, s.applicationBuilder, s.log)
					if err != nil {
						continue
					}
					if rootConfig.ApplicationConfigurations != nil {
						for _, c := range rootConfig.ApplicationConfigurations {
							if c.Name == conf.Name {
								newConf := *c
								if !models.ConfigurationAreEqual(conf, newConf) {
									s.log.Infof(fmt.Sprintf("[APP:%s] Configuration changed", newConf.Name))
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
	s.sesStorage.LoadSessions(s.appStorage, s.sessionBuilder)
}

func (s *Startup) startSessions() {
	for _, session := range s.sesStorage.GetAllAliveSessions() {
		s.mediator.HealthcheckSession.Enqueue(queues.SessionHealthcheckInput{
			Session: session,
		})
	}
}

func (s *Startup) startServer() {
	port := fmt.Sprint(s.configuration.Global.Port)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: s.handler.Router,
	}

	s.log.Infof("Server started on port %s", port)
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
