package services

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/models"
	"gopkg.in/yaml.v2"
)

func LoadConfigurations() (*models.RootConfiguration, *ServiceHandler) {
	dir := getExecutableFolder()

	files := getYamlFiles(dir)

	configurations, serviceHandler := unmarshalConfigurations(files)

	return configurations, serviceHandler

}

func getExecutableFolder() string {
	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalln("Error retrieving file path", err)
	}
	dir := filepath.Dir(executablePath)
	return dir
}

func getYamlFiles(root string) []string {
	files := []string{}

	fileInfos, err := ioutil.ReadDir(root)
	if err != nil {
		log.Fatalln("Error reading from directory", err)
	}
	for _, info := range fileInfos {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".yml") {
			abs, err := filepath.Abs(info.Name())
			if err != nil {
				log.Fatalln("Error resolving configuration file", err)
			}
			files = append(files, abs)
		}
	}

	return files
}

func unmarshalConfigurations(files []string) (*models.RootConfiguration, *ServiceHandler) {
	rootConfiguration := &models.RootConfiguration{
		Services: []*models.Service{},
	}
	for _, file := range files {
		log.Infof("Found configuration file %s", file)
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Errorln(fmt.Sprintf("Could not retrieve content of file %s", file), err)
		}
		var root models.RootConfiguration
		err = yaml.Unmarshal(content, &root)
		if err != nil {
			log.Errorln(fmt.Sprintf("Error in configuration file %s", file), err)
		}
		if root.Global != (models.Global{}) {
			rootConfiguration.Global = root.Global
		}
		if root.Services != nil {
			for _, service := range root.Services {

				builtService, err := models.NewService(service)
				if err != nil {
					log.Errorf("Service %s configuration error: %s", service.Name, err.Error())
				} else {
					rootConfiguration.Services = append(rootConfiguration.Services, builtService)
				}
			}
		}
	}

	serviceHandler := NewServiceHandler(rootConfiguration)

	// Default global configurations
	if rootConfiguration.Global.Port == 0 {
		rootConfiguration.Global.Port = 8888
	}

	if rootConfiguration.Global.SessionsFolder == "" {
		rootConfiguration.Global.SessionsFolder = "./.sessions"
	}

	for _, service := range rootConfiguration.Services {
		err := serviceHandler.InitializeService(service)
		if err != nil {
			log.Fatalln(fmt.Sprintf("Could not provision service %s: %s", service.Name, err.Error()))
			panic(err)
		}
	}

	return rootConfiguration, serviceHandler
}
