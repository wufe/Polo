package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg"
	"github.com/wufe/polo/pkg/background"
	"github.com/wufe/polo/pkg/background/queues"
	"github.com/wufe/polo/pkg/http/proxy"
	"github.com/wufe/polo/pkg/http/rest"
	"github.com/wufe/polo/pkg/http/routing"
	"github.com/wufe/polo/pkg/services"
	"github.com/wufe/polo/pkg/storage"
	"github.com/wufe/polo/pkg/utils"
)

func main() {

	dev := utils.IsDev()
	devServer := utils.DevServerURL()

	// Configuration (.yml)
	configuration := storage.LoadConfigurations()

	// Instance
	existingInstance, _ := storage.DetectInstance()
	if existingInstance == nil {
		storage.NewInstance(fmt.Sprint(configuration.Global.Port)).Persist()
	} else {
		log.Infof("Detected existing instance on host %s", existingInstance.Host)
		return
	}

	// Storage
	database := storage.NewDB()
	appStorage := storage.NewApplication()
	sesStorage := storage.NewSession(database)

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
	background.NewSessionBuildWorker(&configuration.Global, appStorage, sesStorage, mediator)
	background.NewSessionCleanWorker(sesStorage, mediator)
	background.NewSessionFilesystemWorker(mediator)
	background.NewSessionDestroyWorker(mediator)
	background.NewSessionHealthcheckWorker(mediator)
	background.NewApplicationInitWorker(&configuration.Global, mediator)
	background.NewApplicationFetchWorker(sesStorage, mediator)

	// Services
	staticService := services.NewStaticService(dev, devServer)
	queryService := services.NewQueryService(dev, sesStorage, appStorage)
	requestService := services.NewRequestService(dev, sesStorage, appStorage, mediator)

	// HTTP
	proxy := proxy.NewHandler(dev, devServer)
	routing := routing.NewHandler(dev, proxy, sesStorage, appStorage, requestService, staticService)
	rest := rest.NewHandler(dev, staticService, routing, proxy, queryService, requestService)

	// Startup
	pkg.NewStartup(dev, configuration, rest, staticService, appStorage, sesStorage, mediator).Start()

}
