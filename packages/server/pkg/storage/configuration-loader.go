package storage

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/utils"
	"gopkg.in/yaml.v2"
)

func LoadConfigurations(environment utils.Environment, applicationBuilder *models.ApplicationBuilder, logger logging.Logger) (*models.RootConfiguration, []*models.Application) {
	dir := environment.GetExecutableFolder()

	files := getYamlFiles(dir, logger)

	return unmarshalConfigurations(files, applicationBuilder, logger, environment)

}

func getYamlFiles(root string, logger logging.Logger) []string {
	files := []string{}

	fileInfos, err := ioutil.ReadDir(root)
	if err != nil {
		logger.Fatalln("Error reading from directory", err)
	}
	for _, info := range fileInfos {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".yml") {
			abs, err := filepath.Abs(filepath.Join(root, info.Name()))
			if err != nil {
				logger.Fatalln("Error resolving configuration file", err)
			}
			files = append(files, abs)
		}
	}

	return files
}

func unmarshalConfigurations(files []string, applicationBuilder *models.ApplicationBuilder, logger logging.Logger, environment utils.Environment) (*models.RootConfiguration, []*models.Application) {
	rootConfiguration := &models.RootConfiguration{
		ApplicationConfigurations: []*models.ApplicationConfiguration{},
	}
	applications := []*models.Application{}
	for _, file := range files {
		logger.Infof("Found configuration file %s", file)
		root, err := UnmarshalConfiguration(file, applicationBuilder, logger)
		if err != nil {
			continue
		}
		if root.Global != (models.GlobalConfiguration{}) {
			rootConfiguration.Global = root.Global
		}
		if root.ApplicationConfigurations != nil {
			for _, conf := range root.ApplicationConfigurations {

				builtApplication, err := applicationBuilder.Build(conf, file)
				if err != nil {
					logger.Errorf("Application %s configuration error: %s", conf.Name, err.Error())
				} else {
					applications = append(applications, builtApplication)
					conf := builtApplication.GetConfiguration()
					root.ApplicationConfigurations = append(root.ApplicationConfigurations, &conf)
				}
			}
		}
	}

	// Default global configurations
	if rootConfiguration.Global.Port == 0 {
		rootConfiguration.Global.Port = 8888
	}

	if rootConfiguration.Global.SessionsFolder == "" {
		rootConfiguration.Global.SessionsFolder = filepath.Join(environment.GetExecutableFolder(), ".sessions")
	}

	if rootConfiguration.Global.MaxConcurrentSessions == 0 {
		rootConfiguration.Global.MaxConcurrentSessions = 10
	}

	return rootConfiguration, applications
}

func UnmarshalConfiguration(file string, applicationBuilder *models.ApplicationBuilder, logger logging.Logger) (models.RootConfiguration, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		logger.Errorln(fmt.Sprintf("Could not retrieve content of file %s", file), err)
		return models.RootConfiguration{}, err
	}

	var root models.RootConfiguration
	err = yaml.Unmarshal(content, &root)
	if err != nil {
		logger.Errorln(fmt.Sprintf("Error in configuration file %s", file), err)
	}
	if root.ApplicationConfigurations != nil {
		for i, c := range root.ApplicationConfigurations {
			root.ApplicationConfigurations[i], err = applicationBuilder.BuildConfiguration(c)
			if err != nil {
				logger.Errorln(err)
				return root, err
			}
		}
	}

	return root, nil
}
