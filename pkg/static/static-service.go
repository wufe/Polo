package static

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/rakyll/statik/fs"
	"github.com/sasha-s/go-deadlock"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	deadlock.Mutex
	isDev                bool
	devServer            string
	FileSystem           http.FileSystem
	sessionHelperContent string
}

func NewService(isDev bool, devServer string) *Service {
	service := &Service{
		isDev:     isDev,
		devServer: devServer,
	}
	service.initStaticFileSystem()
	return service
}

func (s *Service) SetSessionHelperContent(helper string) {
	s.Lock()
	defer s.Unlock()
	s.sessionHelperContent = helper
}

func (s *Service) GetSessionHelperContent() string {
	s.Lock()
	defer s.Unlock()
	return s.sessionHelperContent
}

func (s *Service) LoadSessionHelper() {
	if s.isDev {
		// If in dev mode, the content is available via webpack dev server
		go func() {
			for {
				resp, err := http.Get(fmt.Sprintf("%s%s%s", s.devServer, "/_polo_/public", "/session-helper.html"))
				if err != nil {
					log.Errorf("Error while getting session helper: %s", err.Error())
				} else {
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Errorf("Error while reading session helper response: %s", err.Error())
					} else {
						s.SetSessionHelperContent(string(body))
					}
					resp.Body.Close()
				}

				time.Sleep(30 * time.Second)
			}
		}()
	} else {
		file, err := s.FileSystem.Open("/session-helper.html")
		if err != nil {
			log.Errorf("Error while getting session helper: %s", err.Error())
		} else {
			defer file.Close()
			content, err := ioutil.ReadAll(file)
			if err != nil {
				log.Errorf("Error while reading session helper content: %s", err.Error())
			} else {
				s.SetSessionHelperContent(string(content))
			}
		}
	}
}

func (s *Service) GetManager() []byte {
	file, err := s.FileSystem.Open("/manager.html")
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("Could not read /manager.html")
		return nil
	}
	return content
}

func (s *Service) initStaticFileSystem() {
	if !s.isDev {
		fileSystem, err := fs.New()
		if err != nil {
			panic(err)
		}
		s.FileSystem = fileSystem
	}
}
