package tests

import (
	"fmt"
	"log"

	"github.com/wufe/polo/internal/tests/storage_fixture"
	"github.com/wufe/polo/internal/tests/utils_fixture"
	"github.com/wufe/polo/pkg"
	"github.com/wufe/polo/pkg/background"
	"github.com/wufe/polo/pkg/background/queues"
	"github.com/wufe/polo/pkg/execution"
	"github.com/wufe/polo/pkg/http/net"
	"github.com/wufe/polo/pkg/http/proxy"
	"github.com/wufe/polo/pkg/http/rest"
	"github.com/wufe/polo/pkg/http/routing"
	"github.com/wufe/polo/pkg/integrations"
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/services"
	"github.com/wufe/polo/pkg/storage"
	"github.com/wufe/polo/pkg/utils"
	"github.com/wufe/polo/pkg/versioning"
	"go.uber.org/dig"
)

type DI struct {
	container  *dig.Container
	injectable *InjectableServices
}

func NewDIContainer(injectable *InjectableServices) *DI {
	return &DI{
		container:  dig.New(),
		injectable: injectable,
	}
}

func (d *DI) GetContainer() *dig.Container {
	return d.container
}

func (d *DI) AddEnvironment() {
	d.container.Provide(func() utils.Environment {
		return utils_fixture.BuildTestEnvironment()
	})
}

func (d *DI) AddLog() {
	if err := d.container.Provide(func(environment utils.Environment) logging.Logger {
		return logging.NewLogger(environment)
	}); err != nil {
		log.Panic(err)
	}
}

// Factories

func (d *DI) AddMutexBuilder() {
	if err := d.container.Provide(func(env utils.Environment) utils.MutexBuilder {
		return func() utils.RWLocker {
			return utils.GetMutex(env)
		}
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddSessionBuilder() {
	if err := d.container.Provide(models.NewSessionBuilder); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddApplicationBuilder() {
	if err := d.container.Provide(models.NewApplicationBuilder); err != nil {
		log.Panic(err)
	}
}

// Git

func (d *DI) AddGitClient() {
	if err := d.container.Provide(func(commandRunner execution.CommandRunner) versioning.GitClient {
		if d.injectable == nil || d.injectable.GitClient == nil {
			return versioning.GetGitClient(commandRunner)
		}
		return d.injectable.GitClient
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddRepositoryFetcher() {
	if err := d.container.Provide(func(gitClient versioning.GitClient) versioning.RepositoryFetcher {
		if d.injectable != nil && d.injectable.RepositoryFetcher != nil {
			return d.injectable.RepositoryFetcher
		}
		return versioning.NewRepositoryFetcher(gitClient)
	}); err != nil {
		log.Panic(err)
	}
}

// Configuration (.yml)

func (d *DI) AddConfiguration(applicationConfigurations ...*models.ApplicationConfiguration) {
	if err := d.container.Provide(func(environment utils.Environment, applicationBuilder *models.ApplicationBuilder) (*models.RootConfiguration, []*models.Application) {
		configuration := &models.RootConfiguration{
			Global: models.GlobalConfiguration{
				SessionsFolder:        environment.GetExecutableFolder() + "/.sessions",
				MaxConcurrentSessions: 999,
			},
			ApplicationConfigurations: applicationConfigurations,
		}

		applications := []*models.Application{}
		for _, conf := range configuration.ApplicationConfigurations {
			application, err := applicationBuilder.Build(conf, "")
			if err != nil {
				panic(err)
			}
			applications = append(applications, application)
		}

		return configuration, applications
	}); err != nil {
		log.Panic(err)
	}
}

// Instance

func (d *DI) AddInstance() {
	if err := d.container.Provide(func(environment utils.Environment, configuration *models.RootConfiguration) (*storage.Instance, error) {
		existingInstance, _ := storage.DetectInstance(environment)
		if existingInstance == nil {
			return nil, fmt.Errorf("Detected existing instance on host %s", existingInstance.Host)
		}
		instance := storage.NewInstance(fmt.Sprint(configuration.Global.Port))
		instance.Persist(environment)
		return instance, nil
	}); err != nil {
		log.Println(err.Error())
	}
}

// Storage

func (d *DI) AddDatabase() {
	if err := d.container.Provide(func(environment utils.Environment, logger logging.Logger) storage.Database {
		return storage_fixture.NewDB(logger)
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddApplicationStorage() {
	if err := d.container.Provide(func(environment utils.Environment) *storage.Application {
		appStorage := storage.NewApplication(environment)
		return appStorage
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddSessionStorage() {
	if err := d.container.Provide(storage.NewSession); err != nil {
		log.Panic(err)
	}
}

// Command

func (d *DI) AddCommandRunner() {
	if err := d.container.Provide(func() execution.CommandRunner {
		if d.injectable != nil && d.injectable.CommandRunner != nil {
			return d.injectable.CommandRunner
		}
		return execution.NewCommandRunner()
	}); err != nil {
		log.Panic(err)
	}
}

// Mediator

func (d *DI) AddSessionBuildQueue() {
	if err := d.container.Provide(func() queues.SessionBuildQueue {
		return queues.NewSessionBuild()
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddSessionDestroyQueue() {
	if err := d.container.Provide(func() queues.SessionDestroyQueue {
		return queues.NewSessionDestroy()
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddSessionFilesystemQueue() {
	if err := d.container.Provide(func() queues.SessionFilesystemQueue {
		return queues.NewSessionFilesystem()
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddSessionCleanupQueue() {
	if err := d.container.Provide(func() queues.SessionCleanupQueue {
		return queues.NewSessionCleanup()
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddSessionStartQueue() {
	if err := d.container.Provide(func() queues.SessionStartQueue {
		return queues.NewSessionStart()
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddSessionHealthCheckQueue() {
	if err := d.container.Provide(func() queues.SessionHealthcheckQueue {
		return queues.NewSessionHealthCheck()
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddApplicationInitQueue() {
	if err := d.container.Provide(func() queues.ApplicationInitQueue {
		return queues.NewApplicationInit()
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddApplicationFetchQueue() {
	if err := d.container.Provide(func() queues.ApplicationFetchQueue {
		return queues.NewApplicationFetch()
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddMediator() {
	if err := d.container.Provide(func(
		sessionBuildQueue queues.SessionBuildQueue,
		sessionDestroyQueue queues.SessionDestroyQueue,
		sessionFilesystemQueue queues.SessionFilesystemQueue,
		sessionCleanupQueue queues.SessionCleanupQueue,
		sessionStartQueue queues.SessionStartQueue,
		sessionHealthcheckQueue queues.SessionHealthcheckQueue,
		applicationInitQueue queues.ApplicationInitQueue,
		applicationFetchQueue queues.ApplicationFetchQueue,
	) *background.Mediator {
		return background.NewMediator(
			sessionBuildQueue,
			sessionDestroyQueue,
			sessionFilesystemQueue,
			sessionCleanupQueue,
			sessionStartQueue,
			sessionHealthcheckQueue,
			applicationInitQueue,
			applicationFetchQueue,
		)
	}); err != nil {
		log.Panic(err)
	}
}

// Workers command execution

func (d *DI) AddSessionCommandExecution() {
	if err := d.container.Provide(background.NewSessionCommandExecution); err != nil {
		log.Panic(err)
	}
}

// Workers

func (d *DI) AddSessionBuildWorker() {
	if err := d.container.Provide(func(
		configuration *models.RootConfiguration,
		appStorage *storage.Application,
		sesStorage *storage.Session,
		mediator *background.Mediator,
		sessionBuilder *models.SessionBuilder,
		logger logging.Logger,
		sessionCommandExecution background.SessionCommandExecution,
		portRetriever net.PortRetriever,
	) *background.SessionBuildWorker {
		return background.NewSessionBuildWorker(&configuration.Global, appStorage, sesStorage, mediator, sessionBuilder, logger, sessionCommandExecution, portRetriever)
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddSessionStartWorker() {
	if err := d.container.Provide(func(sesStorage *storage.Session, mediator *background.Mediator, logger logging.Logger) *background.SessionStartWorker {
		return background.NewSessionStartWorker(sesStorage, mediator, logger)
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddSessionCleanWorker() {
	if err := d.container.Provide(background.NewSessionCleanWorker); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddSessionFilesystemWorker() {
	if err := d.container.Provide(func(gitClient versioning.GitClient, mediator *background.Mediator) *background.SessionFilesystemWorker {
		return background.NewSessionFilesystemWorker(gitClient, mediator)
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddSessionDestroyWorker() {
	if err := d.container.Provide(background.NewSessionDestroyWorker); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddSessionHealthcheckWorker() {
	if err := d.container.Provide(func(mediator *background.Mediator, logger logging.Logger) *background.SessionHealthcheckWorker {
		return background.NewSessionHealthcheckWorker(mediator, logger)
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddApplicationInitWorker() {
	if err := d.container.Provide(func(configuration *models.RootConfiguration, gitClient versioning.GitClient, mediator *background.Mediator, logger logging.Logger) *background.ApplicationInitWorker {
		return background.NewApplicationInitWorker(&configuration.Global, gitClient, mediator, logger)
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddApplicationFetchWorker() {
	if err := d.container.Provide(func(sesStorage *storage.Session, fetcher versioning.RepositoryFetcher, mediator *background.Mediator, logger logging.Logger) *background.ApplicationFetchWorker {
		return background.NewApplicationFetchWorker(sesStorage, fetcher, mediator, logger)
	}); err != nil {
		log.Panic(err)
	}
}

// Services

func (d *DI) AddStaticService() {
	if err := d.container.Provide(services.NewStaticService); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddQueryService() {
	if err := d.container.Provide(func(environment utils.Environment, sesStorage *storage.Session, appStorage *storage.Application, logger logging.Logger) *services.QueryService {
		return services.NewQueryService(environment, sesStorage, appStorage, logger)
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddIntegrationsStatusRetriever() {
	if err := d.container.Provide(func(query *services.QueryService) integrations.SessionIntegrationsStatusRetriever {
		return query
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddRequestService() {
	if err := d.container.Provide(func(environment utils.Environment, sesStorage *storage.Session, appStorage *storage.Application, mediator *background.Mediator) *services.RequestService {
		return services.NewRequestService(environment, sesStorage, appStorage, mediator)
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddAliasingService() {
	if err := d.container.Provide(services.NewAliasingService); err != nil {
		log.Panic(err)
	}
}

// HTTP

func (d *DI) AddPortRetriever() {
	if err := d.container.Provide(func(logger logging.Logger) net.PortRetriever {
		if d.injectable != nil && d.injectable.PortRetriever != nil {
			return d.injectable.PortRetriever
		}
		return net.NewPortRetriever(logger)
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddHTTPProxy() {
	if err := d.container.Provide(func(environment utils.Environment, logger logging.Logger, conf *models.RootConfiguration) *proxy.Handler {
		return proxy.NewHandler(environment, logger, conf)
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddHTTPRouter() {
	if err := d.container.Provide(func(
		environment utils.Environment,
		proxy *proxy.Handler,
		sesStorage *storage.Session,
		appStorage *storage.Application,
		queryService *services.QueryService,
		requestService *services.RequestService,
		staticService *services.StaticService,
		logger logging.Logger,
	) *routing.Handler {
		return routing.NewHandler(environment, proxy, sesStorage, appStorage, queryService, requestService, staticService, logger)
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddHTTPRestHandler() {
	if err := d.container.Provide(func(
		environment utils.Environment,
		staticService *services.StaticService,
		routing *routing.Handler,
		proxy *proxy.Handler,
		queryService *services.QueryService,
		requestService *services.RequestService,
		logger logging.Logger,
	) *rest.Handler {
		return rest.NewHandler(environment, staticService, routing, proxy, queryService, requestService, logger)
	}); err != nil {
		log.Panic(err)
	}
}

func (d *DI) AddIntegrationsHTTPHandler() {
	if err := d.container.Provide(integrations.NewHandler); err != nil {
		log.Panic(err)
	}
}

// Startup

func (d *DI) AddStartup() {
	if err := d.container.Provide(func(params pkg.StartupParams) *pkg.Startup {
		return pkg.NewStartup(params)
	}); err != nil {
		log.Panic(err)
	}
}

// Getters

func (d *DI) GetStartup() *pkg.Startup {
	var startup *pkg.Startup
	if err := d.container.Invoke(func(s *pkg.Startup) {
		startup = s
	}); err != nil {
		log.Panic(err)
	}
	return startup
}

func (d *DI) GetApplications() []*models.Application {
	var applications []*models.Application
	if err := d.container.Invoke(func(a []*models.Application) {
		applications = a
	}); err != nil {
		log.Panic(err)
	}
	return applications
}

func (d *DI) GetMediator() *background.Mediator {
	var mediator *background.Mediator
	if err := d.container.Invoke(func(m *background.Mediator) {
		mediator = m
	}); err != nil {
		log.Panic(err)
	}
	return mediator
}

func (d *DI) GetRequestService() *services.RequestService {
	var service *services.RequestService
	if err := d.container.Invoke(func(s *services.RequestService) {
		service = s
	}); err != nil {
		log.Panic(err)
	}
	return service
}

func (d *DI) GetEnvironment() utils.Environment {
	var environment utils.Environment
	if err := d.container.Invoke(func(e utils.Environment) {
		environment = e
	}); err != nil {
		log.Panic(err)
	}
	return environment
}

func (d *DI) GetSessionStorage() *storage.Session {
	var sessionStorage *storage.Session
	if err := d.container.Invoke(func(s *storage.Session) {
		sessionStorage = s
	}); err != nil {
		log.Panic(err)
	}
	return sessionStorage
}

type InjectableServices struct {
	RepositoryFetcher versioning.RepositoryFetcher
	GitClient         versioning.GitClient
	CommandRunner     execution.CommandRunner
	PortRetriever     net.PortRetriever
}
