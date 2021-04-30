package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
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
	"go.uber.org/dig"
)

func main() {

	c := dig.New()

	c.Provide(func() utils.Environment {
		return utils.DetectEnvironment()
	})

	// Factories

	if err := c.Provide(func(env utils.Environment) utils.MutexBuilder {
		return func() utils.RWLocker {
			return utils.GetMutex(env)
		}
	}); err != nil {
		log.Panic(err)
	}

	if err := c.Provide(communication.NewPubSubBuilder); err != nil {
		log.Panic(err)
	}

	if err := c.Provide(models.NewSessionBuilder); err != nil {
		log.Panic(err)
	}

	if err := c.Provide(models.NewApplicationBuilder); err != nil {
		log.Panic(err)
	}

	// Git dependencies

	if err := c.Provide(func() fetch.RepositoryFetcher {
		return fetch.NewRepositoryFetcher()
	}); err != nil {
		log.Panic(err)
	}

	// Configuration (.yml)

	if err := c.Provide(func(environment utils.Environment, applicationBuilder *models.ApplicationBuilder) (*models.RootConfiguration, []*models.Application) {
		return storage.LoadConfigurations(environment, applicationBuilder)
	}); err != nil {
		log.Panic(err)
	}

	// Instance

	if err := c.Provide(func(environment utils.Environment, configuration *models.RootConfiguration) (*storage.Instance, error) {
		existingInstance, _ := storage.DetectInstance(environment)
		if existingInstance == nil {
			return nil, fmt.Errorf("Detected existing instance on host %s", existingInstance.Host)
		}
		instance := storage.NewInstance(fmt.Sprint(configuration.Global.Port))
		instance.Persist(environment)
		return instance, nil
	}); err != nil {
		log.Infof(err.Error())
	}

	// Storage

	if err := c.Provide(func(environment utils.Environment) *storage.Database {
		return storage.NewDB(environment.GetExecutableFolder())
	}); err != nil {
		log.Panic(err)
	}

	if err := c.Provide(func(environment utils.Environment, database *storage.Database) (*storage.Application, *storage.Session) {
		appStorage := storage.NewApplication(environment)
		sesStorage := storage.NewSession(database, environment)
		return appStorage, sesStorage
	}); err != nil {
		log.Panic(err)
	}

	if err := c.Provide(func() *background.Mediator {
		return background.NewMediator(
			queues.NewSessionBuild(),
			queues.NewSessionDestroy(),
			queues.NewSessionFilesystem(),
			queues.NewSessionCleanup(),
			queues.NewSessionStart(),
			queues.NewSessionHealthCheck(),
			queues.NewApplicationInit(),
			queues.NewApplicationFetch(),
		)
	}); err != nil {
		log.Panic(err)
	}

	// Workers

	if err := c.Provide(func(
		configuration *models.RootConfiguration,
		appStorage *storage.Application,
		sesStorage *storage.Session,
		mediator *background.Mediator,
		sessionBuilder *models.SessionBuilder,
		pubSubBuilder *communication.PubSubBuilder,
	) *background.SessionBuildWorker {
		return background.NewSessionBuildWorker(&configuration.Global, appStorage, sesStorage, mediator, sessionBuilder, pubSubBuilder)
	}); err != nil {
		log.Panic(err)
	}

	if err := c.Provide(func(sesStorage *storage.Session, mediator *background.Mediator) *background.SessionStartWorker {
		return background.NewSessionStartWorker(sesStorage, mediator)
	}); err != nil {
		log.Panic(err)
	}

	if err := c.Provide(func(sesStorage *storage.Session, mediator *background.Mediator) *background.SessionCleanWorker {
		return background.NewSessionCleanWorker(sesStorage, mediator)
	}); err != nil {
		log.Panic(err)
	}

	if err := c.Provide(func(mediator *background.Mediator) *background.SessionFilesystemWorker {
		return background.NewSessionFilesystemWorker(mediator)
	}); err != nil {
		log.Panic(err)
	}

	if err := c.Provide(func(mediator *background.Mediator) *background.SessionDestroyWorker {
		return background.NewSessionDestroyWorker(mediator)
	}); err != nil {
		log.Panic(err)
	}

	if err := c.Provide(func(mediator *background.Mediator) *background.SessionHealthcheckWorker {
		return background.NewSessionHealthcheckWorker(mediator)
	}); err != nil {
		log.Panic(err)
	}

	if err := c.Provide(func(configuration *models.RootConfiguration, mediator *background.Mediator) *background.ApplicationInitWorker {
		return background.NewApplicationInitWorker(&configuration.Global, mediator)
	}); err != nil {
		log.Panic(err)
	}

	if err := c.Provide(func(sesStorage *storage.Session, fetcher fetch.RepositoryFetcher, mediator *background.Mediator) *background.ApplicationFetchWorker {
		return background.NewApplicationFetchWorker(sesStorage, fetcher, mediator)
	}); err != nil {
		log.Panic(err)
	}

	// Services

	if err := c.Provide(func(environment utils.Environment) *services.StaticService {
		return services.NewStaticService(environment)
	}); err != nil {
		log.Panic(err)
	}

	if err := c.Provide(func(environment utils.Environment, sesStorage *storage.Session, appStorage *storage.Application) *services.QueryService {
		return services.NewQueryService(environment, sesStorage, appStorage)
	}); err != nil {
		log.Panic(err)
	}

	if err := c.Provide(func(environment utils.Environment, sesStorage *storage.Session, appStorage *storage.Application, mediator *background.Mediator) *services.RequestService {
		return services.NewRequestService(environment, sesStorage, appStorage, mediator)
	}); err != nil {
		log.Panic(err)
	}

	// HTTP

	if err := c.Provide(func(environment utils.Environment) *proxy.Handler {
		return proxy.NewHandler(environment)
	}); err != nil {
		log.Panic(err)
	}

	if err := c.Provide(func(
		environment utils.Environment,
		proxy *proxy.Handler,
		sesStorage *storage.Session,
		appStorage *storage.Application,
		queryService *services.QueryService,
		requestService *services.RequestService,
		staticService *services.StaticService,
	) *routing.Handler {
		return routing.NewHandler(environment, proxy, sesStorage, appStorage, queryService, requestService, staticService)
	}); err != nil {
		log.Panic(err)
	}

	if err := c.Provide(func(
		environment utils.Environment,
		staticService *services.StaticService,
		routing *routing.Handler,
		proxy *proxy.Handler,
		queryService *services.QueryService,
		requestService *services.RequestService,
	) *rest.Handler {
		return rest.NewHandler(environment, staticService, routing, proxy, queryService, requestService)
	}); err != nil {
		log.Panic(err)
	}

	// Startup

	if err := c.Provide(func(
		configuration *models.RootConfiguration,
		applications []*models.Application,
		rest *rest.Handler,
		staticService *services.StaticService,
		appStorage *storage.Application,
		sesStorage *storage.Session,
		mediator *background.Mediator,
		applicationBuilder *models.ApplicationBuilder,
		sessionBuilder *models.SessionBuilder,
	) *pkg.Startup {
		return pkg.NewStartup(
			configuration,
			applications,
			rest,
			staticService,
			appStorage,
			sesStorage,
			mediator,
			applicationBuilder,
			sessionBuilder,
		)
	}); err != nil {
		log.Panic(err)
	}

	if err := c.Invoke(func(startup *pkg.Startup) {
		startup.Start(&pkg.StartupOptions{
			WatchApplications: true,
			LoadSessionHelper: true,
			StartServer:       true,
		})
	}); err != nil {
		log.Panic(err)
	}
}
