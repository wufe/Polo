package tests

import (
	"github.com/wufe/polo/internal/tests/storage_fixture"
	"github.com/wufe/polo/internal/tests/utils_fixture"
	"github.com/wufe/polo/pkg"
	"github.com/wufe/polo/pkg/background"
	"github.com/wufe/polo/pkg/background/communication"
	"github.com/wufe/polo/pkg/background/fetch"
	"github.com/wufe/polo/pkg/background/queues"
	"github.com/wufe/polo/pkg/http/proxy"
	"github.com/wufe/polo/pkg/http/rest"
	"github.com/wufe/polo/pkg/http/routing"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/services"
	"github.com/wufe/polo/pkg/storage"
	"github.com/wufe/polo/pkg/utils"
)

type InjectableServices struct {
	RepositoryFetcher fetch.RepositoryFetcher
}

func (s *InjectableServices) GetRepositoryFetcher() fetch.RepositoryFetcher {
	if s == nil || s.RepositoryFetcher == nil {
		return fetch.NewRepositoryFetcher()
	}
	return s.RepositoryFetcher
}

func Fixture(applicationConfiguration *models.ApplicationConfiguration, injectable *InjectableServices) []*models.Application {

	environment := utils_fixture.BuildTestEnvironment()

	configuration := &models.RootConfiguration{
		Global: models.GlobalConfiguration{
			SessionsFolder: environment.GetExecutableFolder() + "/.sessions",
		},
		ApplicationConfigurations: []*models.ApplicationConfiguration{
			applicationConfiguration,
		},
	}

	// Factories
	var mutexBuilder utils.MutexBuilder = func() utils.RWLocker { return utils.GetMutex(environment) }
	pubSubBuilder := communication.NewPubSubBuilder(mutexBuilder)
	sessionBuilder := models.NewSessionBuilder(mutexBuilder, pubSubBuilder)
	applicationBuilder := models.NewApplicationBuilder(mutexBuilder, pubSubBuilder)

	applications := []*models.Application{}

	for _, conf := range configuration.ApplicationConfigurations {
		application, err := applicationBuilder.Build(conf, "")
		if err != nil {
			panic(err)
		}
		applications = append(applications, application)
	}

	// Git dependencies
	fetcher := injectable.GetRepositoryFetcher()

	// Storage
	database := storage_fixture.NewDB(environment.GetExecutableFolder(), &storage_fixture.FixtureDBOptions{
		Clean: true,
	})
	appStorage := storage.NewApplication(environment)
	sesStorage := storage.NewSession(database, environment)

	mediator := background.NewMediator(
		queues.NewSessionBuild(),
		queues.NewSessionDestroy(),
		queues.NewSessionFilesystem(),
		queues.NewSessionCleanup(),
		queues.NewSessionStart(),
		queues.NewSessionHealthCheck(),
		queues.NewApplicationInit(),
		queues.NewApplicationFetch(),
	)

	// Workers
	background.NewSessionBuildWorker(&configuration.Global, appStorage, sesStorage, mediator, sessionBuilder, pubSubBuilder)
	background.NewSessionStartWorker(sesStorage, mediator)
	background.NewSessionCleanWorker(sesStorage, mediator)
	background.NewSessionFilesystemWorker(mediator)
	background.NewSessionDestroyWorker(mediator)
	background.NewSessionHealthcheckWorker(mediator)
	background.NewApplicationInitWorker(&configuration.Global, mediator)
	background.NewApplicationFetchWorker(sesStorage, fetcher, mediator)

	// Services
	staticService := services.NewStaticService(environment)
	queryService := services.NewQueryService(environment, sesStorage, appStorage)
	requestService := services.NewRequestService(environment, sesStorage, appStorage, mediator)

	// HTTP
	proxy := proxy.NewHandler(environment)
	routing := routing.NewHandler(environment, proxy, sesStorage, appStorage, queryService, requestService, staticService)
	rest := rest.NewHandler(environment, staticService, routing, proxy, queryService, requestService)

	// Startup
	pkg.NewStartup(
		configuration,
		applications,
		rest,
		staticService,
		appStorage,
		sesStorage,
		mediator,
		applicationBuilder,
		sessionBuilder,
	).Start(&pkg.StartupOptions{
		WatchApplications: false,
		LoadSessionHelper: false,
		StartServer:       false,
	})

	return applications
}
