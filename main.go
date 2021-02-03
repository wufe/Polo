package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/net"
	"github.com/wufe/polo/services"
)

func main() {

	var wg sync.WaitGroup

	db, err := services.StartDB()
	if err != nil {
		log.Fatal("Cannot create database: " + err.Error())
		return
	}
	defer db.Close()

	configuration, serviceHandler := services.LoadConfigurations()

	sessionHandler := services.NewSessionHandler(configuration, serviceHandler, db)

	port := flag.String("port", fmt.Sprint(configuration.Global.Port), "Port")
	flag.Parse()

	net.NewHTTPServer(*port, sessionHandler, configuration, &wg)
	wg.Wait()

}

func getExecutableFolder() string {
	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalln("Error retrieving file path", err)
	}
	dir := filepath.Dir(executablePath)
	return dir
}
