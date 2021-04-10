package tests

import (
	"github.com/wufe/polo/internal/tests/storage_fixture"
	"github.com/wufe/polo/internal/tests/utils_fixture"
	"github.com/wufe/polo/pkg"
	"github.com/wufe/polo/pkg/background"
	"github.com/wufe/polo/pkg/background/queues"
	"github.com/wufe/polo/pkg/http/proxy"
	"github.com/wufe/polo/pkg/http/rest"
	"github.com/wufe/polo/pkg/http/routing"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/services"
	"github.com/wufe/polo/pkg/storage"
)

func Fixture() {
	environment := utils_fixture.BuildTestEnvironment()

	configuration := &models.RootConfiguration{}
	applications := []*models.Application{}

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
	background.NewSessionBuildWorker(&configuration.Global, appStorage, sesStorage, mediator, environment)
	background.NewSessionStartWorker(sesStorage, mediator)
	background.NewSessionCleanWorker(sesStorage, mediator)
	background.NewSessionFilesystemWorker(mediator)
	background.NewSessionDestroyWorker(mediator)
	background.NewSessionHealthcheckWorker(mediator)
	background.NewApplicationInitWorker(&configuration.Global, mediator)
	background.NewApplicationFetchWorker(sesStorage, mediator)

	// Services
	staticService := services.NewStaticService(environment)
	queryService := services.NewQueryService(environment, sesStorage, appStorage)
	requestService := services.NewRequestService(environment, sesStorage, appStorage, mediator)

	// HTTP
	proxy := proxy.NewHandler(environment)
	routing := routing.NewHandler(environment, proxy, sesStorage, appStorage, queryService, requestService, staticService)
	rest := rest.NewHandler(environment, staticService, routing, proxy, queryService, requestService)

	// Startup
	pkg.NewStartup(environment, configuration, applications, rest, staticService, appStorage, sesStorage, mediator).Start()
}
