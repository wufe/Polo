package models

import (
	_ "github.com/sirupsen/logrus"
)

type RootConfiguration struct {
	Global   Global
	Services []*Service
}

type Global struct {
	Port           int
	TLSCertFile    string `yaml:"tls_cert,omitempty"`
	TLSKeyFile     string `yaml:"tls_key,omitempty"`
	SessionsFolder string `yaml:"sessions_folder"`
}

type Headers struct {
	Add []string `json:"add"`
}

type Healthcheck struct {
	Method       string `json:"method"`
	URL          string `yaml:"url" json:"url"`
	Status       int    `json:"status"`
	RetryTimeout int    `yaml:"retry_timeout" json:"retryTimeout"`
}

type Recycle struct {
	InactivityTimeout int `yaml:"inactivity_timeout" json:"inactivityTimeout"`
}

type Commands struct {
	Start []Command `json:"start"`
	Stop  []Command `json:"stop"`
}

type Command struct {
	Command     string   `json:"command"`
	Environment []string `yaml:"environment,omitempty" json:"environment"`
}

type PortConfiguration struct {
	Except []int `json:"except"`
}
