package main

import (
	"github.com/wufe/polo/pkg"
)

func main() {

	container := pkg.NewDIContainer()

	// Environment

	container.AddEnvironment()

	// Logs

	container.AddLog()

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

	// Command

	container.AddCommandRunner()

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

	// Workers command execution

	container.AddSessionCommandExecution()

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
	container.AddAliasingService()

	// HTTP

	container.AddPortRetriever()
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
