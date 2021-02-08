package main

import (
	"github.com/wufe/polo/pkg/background"
	"github.com/wufe/polo/pkg/background/pipe"
	"github.com/wufe/polo/pkg/http/proxy"
	"github.com/wufe/polo/pkg/http/rest"
	"github.com/wufe/polo/pkg/http/routing"
	"github.com/wufe/polo/pkg/query"
	"github.com/wufe/polo/pkg/request"
	"github.com/wufe/polo/pkg/startup"
	"github.com/wufe/polo/pkg/static"
	"github.com/wufe/polo/pkg/storage"
	"github.com/wufe/polo/pkg/utils"
)

func main() {

	dev := utils.IsDev()
	devServer := utils.DevServerURL()

	// Configuration (.yml)
	configuration := storage.LoadConfigurations()

	// Storage
	database := storage.NewDB()
	appStorage := storage.NewApplication()
	sesStorage := storage.NewSession(database)

	mediator := background.NewMediator(
		pipe.NewSessionBuild(),
		pipe.NewSessionDestroy(),
		pipe.NewSessionFilesystem(),
		pipe.NewSessionCleanup(),
		pipe.NewSessionStart(),
		pipe.NewApplicationInit(),
		pipe.NewApplicationFetch(),
	)

	// Workers
	background.NewSessionBuildWorker(&configuration.Global, appStorage, sesStorage, mediator)
	background.NewSessionCleanWorker(sesStorage, mediator)
	background.NewSessionFilesystemWorker(mediator)
	background.NewSessionDestroyWorker(mediator)
	background.NewApplicationInitWorker(&configuration.Global, mediator)
	background.NewApplicationFetchWorker(mediator)

	// Services
	staticService := static.NewService(dev, devServer)
	queryService := query.NewService(dev, sesStorage, appStorage)
	requestService := request.NewRequestService(dev, sesStorage, appStorage, mediator)

	// HTTP
	proxy := proxy.NewHandler(dev, devServer)
	routing := routing.NewHandler(dev, proxy, sesStorage, appStorage, requestService, staticService)
	rest := rest.NewHandler(dev, staticService, routing, proxy, queryService, requestService)

	// Startup
	startup.NewService(dev, configuration, rest, staticService, appStorage, sesStorage, mediator).Start()

}
