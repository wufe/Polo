package models

import (
	_ "github.com/sirupsen/logrus"
)

type RootConfiguration struct {
	Global       GlobalConfiguration
	Applications []*Application
}

type GlobalConfiguration struct {
	Port                  int
	TLSCertFile           string `yaml:"tls_cert,omitempty"`
	TLSKeyFile            string `yaml:"tls_key,omitempty"`
	SessionsFolder        string `yaml:"sessions_folder"`
	MaxConcurrentSessions int    `yaml:"max_concurrent_sessions" json:"maxConcurrentSessions"`
}

type Headers struct {
	Add    []string `json:"add"`
	Remove []string `json:"remove"`
}

type Healthcheck struct {
	Method        string `json:"method"`
	URL           string `yaml:"url" json:"url"`
	Status        int    `json:"status"`
	RetryInterval int    `yaml:"retry_interval" json:"retryInterval"`
	RetryTimeout  int    `yaml:"retry_timeout" json:"retryTimeout"`
}

type Recycle struct {
	InactivityTimeout int `yaml:"inactivity_timeout" json:"inactivityTimeout"`
}

type Commands struct {
	Start []Command `json:"start"`
	Stop  []Command `json:"stop"`
}

type Command struct {
	Command             string   `json:"command"`
	Environment         []string `yaml:"environment,omitempty" json:"environment"`
	OutputVariable      string   `yaml:"output_variable,omitempty" json:"outputVariable"`
	ContinueOnError     bool     `yaml:"continue_on_error" json:"continueOnError"`
	WorkingDir          string   `yaml:"working_dir" json:"workingDir"`
	StartHealthchecking bool     `yaml:"start_healthchecking" json:"startHealthchecking"`
}

type PortConfiguration struct {
	Except []int `json:"except"`
}
