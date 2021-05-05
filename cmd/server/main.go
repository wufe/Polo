package main

import (
	"github.com/wufe/polo/pkg"
)

func main() {

	pkg.ConfigureLogging()

	container := pkg.NewDIContainer()

	// Environment

	container.AddEnvironment()

	// Factories

	container.AddMutexBuilder()
	container.AddSessionBuilder()
	container.AddApplicationBuilder()

	// Git

	container.AddGitClient()
	container.AddRepositoryFetcher()

	// Configuration

	container.AddConfiguration()

	// Instance

	container.AddInstance()

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
		WatchApplications: true,
		LoadSessionHelper: true,
		StartServer:       true,
	})
}
