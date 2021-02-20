package models

import "github.com/wufe/polo/pkg/utils"

type ApplicationConfiguration struct {
	utils.RWLocker        `json:"-"`
	Name                  string            `json:"name"`
	Remote                string            `json:"remote"`
	Target                string            `json:"target"`
	Host                  string            `json:"host"`
	Fetch                 Fetch             `json:"fetch"`
	Watch                 Watch             `json:"watch"`
	IsDefault             bool              `yaml:"is_default" json:"isDefault"`
	Forwards              []Forward         `json:"forwards"`
	Headers               Headers           `json:"headers"`
	Healthcheck           Healthcheck       `json:"healthCheck"`
	Startup               Startup           `json:"startup"`
	Recycle               Recycle           `json:"recycle"`
	Commands              Commands          `json:"commands"`
	MaxConcurrentSessions int               `yaml:"max_concurrent_sessions" json:"maxConcurrentSessions"`
	Port                  PortConfiguration `yaml:"port" json:"port"`
	UseFolderCopy         bool              `yaml:"use_folder_copy" json:"useFolderCopy"`
	CleanOnExit           *bool             `yaml:"clean_on_exit" json:"cleanOnExit" default:"true"`
}

type Startup struct {
	Timeout int `json:"timeout"`
	Retries int `json:"retries"`
}

type Forward struct {
	Pattern string  `json:"pattern"`
	To      string  `json:"to"`
	Host    string  `json:"host"`
	Headers Headers `json:"headers"`
}

type Watch []string

func (w *Watch) Contains(obj string) bool {
	for _, o := range *w {
		if o == obj {
			return true
		}
	}
	return false
}

type Fetch struct {
	Interval int `json:"interval"`
}

func (a *ApplicationConfiguration) WithLock(f func(*ApplicationConfiguration)) {
	a.Lock()
	defer a.Unlock()
	f(a)
}

func (a *ApplicationConfiguration) WithRLock(f func(*ApplicationConfiguration)) {
	a.RLock()
	defer a.RUnlock()
	f(a)
}
