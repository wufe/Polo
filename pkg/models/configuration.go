package models

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	_ "github.com/sirupsen/logrus"
)

var (
	ErrMalformedHeader error = errors.New("Malformed header; the format should be key=value")
)

type RootConfiguration struct {
	Global                    GlobalConfiguration
	ApplicationConfigurations []*ApplicationConfiguration `yaml:"applications"`
}

type GlobalConfiguration struct {
	Port                  int
	TLSCertFile           string `yaml:"tls_cert,omitempty"`
	TLSKeyFile            string `yaml:"tls_key,omitempty"`
	SessionsFolder        string `yaml:"sessions_folder"`
	MaxConcurrentSessions int    `yaml:"max_concurrent_sessions" json:"maxConcurrentSessions"`
}

type Header string

func (h Header) Parse() (string, string, error) {
	kv := strings.Split(fmt.Sprint(h), "=")
	if len(kv) != 2 {
		return "", "", ErrMalformedHeader
	}
	return kv[0], kv[1], nil
}

type Headers struct {
	Add     []Header `json:"add"`
	Set     []Header `json:"set"`
	Del     []string `json:"del"`
	Replace []Header `json:"replace"`
}

func (h *Headers) ApplyTo(r *http.Request) error {
	var err error
	var k string
	var v string

	for _, header := range h.Replace {
		k, v, err = header.Parse()
		if err == nil {
			if o := r.Header.Get(k); o != "" {
				r.Header.Set(k, v)
			}
		}
	}

	for _, header := range h.Add {
		k, v, err = header.Parse()
		if err == nil {
			r.Header.Add(k, v)
		}
	}

	for _, header := range h.Set {
		k, v, err = header.Parse()
		if err == nil {
			r.Header.Set(k, v)
		}
	}

	for _, header := range h.Del {
		r.Header.Del(header)
	}

	return err
}

type Healthcheck struct {
	Method        string `json:"method"`
	URL           string `yaml:"url" json:"url"`
	Status        int    `json:"status"`
	MaxRetries    int    `yaml:"max_retries" json:"maxRetries"`
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
	Timeout             int      `json:"timeout"`
}

type PortConfiguration struct {
	Except []int `json:"except"`
}
