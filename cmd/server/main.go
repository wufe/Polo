package main

import (
	"github.com/wufe/polo/background"
	"github.com/wufe/polo/background/pipe"
	"github.com/wufe/polo/http/proxy"
	"github.com/wufe/polo/http/rest"
	"github.com/wufe/polo/http/routing"
	"github.com/wufe/polo/query"
	"github.com/wufe/polo/request"
	"github.com/wufe/polo/services"
	"github.com/wufe/polo/startup"
	"github.com/wufe/polo/static"
	"github.com/wufe/polo/storage"
	"github.com/wufe/polo/utils"
)

func main() {

	// var wg sync.WaitGroup

	// db, err := services.StartDB()
	// if err != nil {
	// 	log.Fatal("Cannot create database: " + err.Error())
	// 	return
	// }
	// defer db.Close()

	// OLD
	configuration := services.LoadConfigurations()
	// // sessionHandler := services.NewSessionHandler(configuration, applicationHandler/*, db*/)
	// sessionHandler := new(services.SessionHandler)

	dev := utils.IsDev()
	devServer := utils.DevServerURL()

	// Storage
	appStorage := storage.NewApplication()
	sesStorage := storage.NewSession()

	mediator := background.NewMediator(
		pipe.NewSessionBuild(),
		pipe.NewSessionDestroy(),
		pipe.NewSessionFilesystem(),
		pipe.NewSessionCleanup(),
		pipe.NewApplicationInit(),
		pipe.NewApplicationFetch(),
	)

	// Workers
	background.NewSessionBuildWorker(&configuration.Global, appStorage, sesStorage, mediator)
	background.NewSessionCleanWorker(mediator)
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
	startup.NewService(dev, configuration, rest, staticService, appStorage, mediator).Start()

}
