package main

import (
	"flag"
	"fmt"
	"sync"

	"github.com/wufe/polo/net"
	"github.com/wufe/polo/services"
)

func main() {

	configuration, serviceHandler := services.LoadConfigurations()

	sessionHandler := services.NewSessionHandler(configuration, serviceHandler)

	port := flag.String("port", fmt.Sprint(configuration.Global.Port), "Port")
	flag.Parse()

	var wg sync.WaitGroup

	net.NewHTTPServer(*port, sessionHandler, configuration, &wg)
	wg.Wait()
}
