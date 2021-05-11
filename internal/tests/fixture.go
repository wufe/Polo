package tests

import (
	"github.com/wufe/polo/pkg"
	"github.com/wufe/polo/pkg/models"
)

func Fixture(applicationConfiguration *models.ApplicationConfiguration, injectable *InjectableServices) *DI {

	container := NewDIContainer(injectable)

	// Environment

	container.AddEnvironment()

	// Logging

	pkg.ConfigureLogging(container)

	// Factories

	container.AddMutexBuilder()
	container.AddSessionBuilder()
	container.AddApplicationBuilder()

	// Git

	container.AddGitClient()
	container.AddRepositoryFetcher()

	// Configuration

	container.AddConfiguration(applicationConfiguration)

	// Storage

	container.AddDatabase()
	container.AddApplicationStorage()
	container.AddSessionStorage()

	// Mediator

	container.AddSessionBuildQueue()
	container.AddSessionDestroyQueue()
	container.AddSessionFilesystemQueue()
	container.AddSessionCleanupQueue()
	container.AddSessionStartQueue()
	container.AddSessionHealthCheckQueue()
	container.AddApplicationInitQueue()
	container.AddApplicationFetchQueue()
	container.AddMediator()

	// Workers

	container.AddSessionBuildWorker()
	container.AddSessionStartWorker()
	container.AddSessionCleanWorker()
	container.AddSessionFilesystemWorker()
	container.AddSessionDestroyWorker()
	container.AddSessionHealthcheckWorker()
	container.AddApplicationInitWorker()
	container.AddApplicationFetchWorker()

	// Services

	container.AddStaticService()
	container.AddQueryService()
	container.AddRequestService()

	// HTTP

	container.AddHTTPProxy()
	container.AddHTTPRouter()
	container.AddHTTPRestHandler()

	// Startup

	container.AddStartup()

	container.GetStartup().Start(&pkg.StartupOptions{
		WatchApplications: false,
		LoadSessionHelper: false,
		StartServer:       false,
	})

	return container
}
