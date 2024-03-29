package services

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/utils"
)

//go:embed static/*
var static embed.FS

type StaticService struct {
	utils.RWLocker
	isDev                bool
	devServer            string
	FileSystem           fs.FS
	sessionHelperContent string
	log                  logging.Logger
	configuration        *models.RootConfiguration
}

func NewStaticService(
	environment utils.Environment,
	logger logging.Logger,
	configuration *models.RootConfiguration,
) *StaticService {
	service := &StaticService{
		RWLocker:      utils.GetMutex(environment),
		isDev:         environment.IsDev(),
		devServer:     environment.DevServerURL(),
		log:           logger,
		configuration: configuration,
	}
	service.initStaticFileSystem()
	return service
}

func (s *StaticService) SetSessionHelperContent(helper string) {
	s.Lock()
	defer s.Unlock()
	s.sessionHelperContent = helper
}

func (s *StaticService) GetSessionHelperContent() string {
	s.RLock()
	defer s.RUnlock()
	return s.sessionHelperContent
}

func (s *StaticService) LoadSessionHelper() {
	if s.isDev {
		// If in dev mode, the content is available via webpack dev server
		go func() {
			for {
				resp, err := http.Get(fmt.Sprintf("%s%s%s", s.devServer, "/_polo_/public", "/session-helper.html"))
				if err != nil {
					s.log.Errorf("Error while getting session helper: %s", err.Error())
				} else {
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						s.log.Errorf("Error while reading session helper response: %s", err.Error())
					} else {
						s.SetSessionHelperContent(string(body))
					}
					resp.Body.Close()
				}

				time.Sleep(30 * time.Second)
			}
		}()
	} else {
		file, err := s.FileSystem.Open("session-helper.html")
		if err != nil {
			s.log.Errorf("Error while getting session helper: %s", err.Error())
		} else {
			defer file.Close()
			content, err := ioutil.ReadAll(file)
			if err != nil {
				s.log.Errorf("Error while reading session helper content: %s", err.Error())
			} else {
				s.SetSessionHelperContent(string(content))
			}
		}
	}
}

func (s *StaticService) GetManager() []byte {
	file, err := s.FileSystem.Open("manager.html")
	if err != nil {
		s.log.Errorf("error opening file: %s", err)
		return nil
	}
	content, err := io.ReadAll(file)
	if err != nil {
		s.log.Errorf("Could not read manager.html")
		return nil
	}

	serializedConfiguration, err := json.Marshal(s.configuration.GetManagerConfiguration())
	if err != nil {
		s.log.Errorf("error serializing configuration: %s", err)
		return nil
	}
	content = []byte(strings.ReplaceAll(string(content), "{}", string(serializedConfiguration)))

	return content
}

func (s *StaticService) initStaticFileSystem() {
	if !s.isDev {
		var err error
		s.FileSystem, err = fs.Sub(static, "static")
		if err != nil {
			panic(err)
		}
	}
}
