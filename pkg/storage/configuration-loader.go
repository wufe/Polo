package storage

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/utils"
	"gopkg.in/yaml.v2"
)

func LoadConfigurations() (*models.RootConfiguration, []*models.Application) {
	dir := utils.GetExecutableFolder()

	files := getYamlFiles(dir)

	return unmarshalConfigurations(files)

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

func unmarshalConfigurations(files []string) (*models.RootConfiguration, []*models.Application) {
	rootConfiguration := &models.RootConfiguration{
		ApplicationConfigurations: []*models.ApplicationConfiguration{},
	}
	applications := []*models.Application{}
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
		if root.Global != (models.GlobalConfiguration{}) {
			rootConfiguration.Global = root.Global
		}
		if root.ApplicationConfigurations != nil {
			for _, conf := range root.ApplicationConfigurations {

				builtApplication, err := models.NewApplication(conf)
				if err != nil {
					log.Errorf("Application %s configuration error: %s", conf.Name, err.Error())
				} else {
					applications = append(applications, builtApplication)
					rootConfiguration.ApplicationConfigurations = append(rootConfiguration.ApplicationConfigurations, &builtApplication.Configuration)
				}
			}
		}
	}

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

	return rootConfiguration, applications
}
