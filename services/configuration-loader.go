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

func LoadConfigurations() (*models.RootConfiguration, *ApplicationHandler) {
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

func unmarshalConfigurations(files []string) (*models.RootConfiguration, *ApplicationHandler) {
	rootConfiguration := &models.RootConfiguration{
		Applications: []*models.Application{},
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
		if root.Applications != nil {
			for _, application := range root.Applications {

				builtApplication, err := models.NewApplication(application)
				if err != nil {
					log.Errorf("Application %s configuration error: %s", application.Name, err.Error())
				} else {
					rootConfiguration.Applications = append(rootConfiguration.Applications, builtApplication)
				}
			}
		}
	}

	applicationHandler := NewApplicationHandler(rootConfiguration)

	// Default global configurations
	if rootConfiguration.Global.Port == 0 {
		rootConfiguration.Global.Port = 8888
	}

	if rootConfiguration.Global.SessionsFolder == "" {
		rootConfiguration.Global.SessionsFolder = "./.sessions"
	}

	if rootConfiguration.Global.MaxConcurrentSessions == 0 {
		rootConfiguration.Global.MaxConcurrentSessions = 10
	}

	for _, application := range rootConfiguration.Applications {
		err := applicationHandler.InitializeApplication(application)
		if err != nil {
			log.Fatalln(fmt.Sprintf("Could not provision application %s: %s", application.Name, err.Error()))
			panic(err)
		}
	}

	return rootConfiguration, applicationHandler
}
