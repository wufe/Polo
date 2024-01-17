package models

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrMalformedHeader error = errors.New("Malformed header; the format should be key=value")
)

type RootConfiguration struct {
	Global                    GlobalConfiguration
	ApplicationConfigurations []*ApplicationConfiguration `yaml:"applications"`
}

type GlobalConfiguration struct {
	Host                  string
	Port                  int
	PublicURL             string                       `yaml:"public_url,omitempty"`
	Debug                 bool                         `yaml:"debug,omitempty"`
	TLSCertFile           string                       `yaml:"tls_cert,omitempty"`
	TLSKeyFile            string                       `yaml:"tls_key,omitempty"`
	SessionsFolder        string                       `yaml:"sessions_folder"`
	MaxConcurrentSessions int                          `yaml:"max_concurrent_sessions" json:"maxConcurrentSessions"`
	FeaturesPreview       FeaturesPreviewConfiguration `yaml:"features_preview" json:"featuresPreview"`
	Integrations          IntegrationsConfiguration    `yaml:"integrations" json:"integrations"`
}

type IntegrationsConfiguration struct {
	Enabled bool                            `yaml:"enabled" json:"enabled"`
	Server  IntegrationsServerConfiguration `yaml:"server" json:"server"`
	Tilt    TiltConfiguration               `yaml:"tilt" json:"tilt"`
}

type IntegrationsServerConfiguration struct {
	Host      string `yaml:"host" json:"host"`
	Port      int    `yaml:"port" json:"port"`
	PublicURL string `yaml:"public_url" json:"publicURL"`
}

type TiltConfiguration struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type FeaturesPreviewConfiguration struct {
	AdvancedTerminalOutput bool `yaml:"advanced_terminal_output"`
}

// ManagerConfiguration is the struct containing all the configuration
// pieced which will be serialized and sent to the manager
type ManagerConfiguration struct {
	AdvancedTerminalOutput bool   `json:"advancedTerminalOutput"`
	IntegrationsPublicURL  string `json:"integrationsPublicURL"`
}

func (c *RootConfiguration) GetManagerConfiguration() ManagerConfiguration {
	return ManagerConfiguration{
		AdvancedTerminalOutput: c.Global.FeaturesPreview.AdvancedTerminalOutput,
		IntegrationsPublicURL:  c.Global.Integrations.Server.PublicURL,
	}
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
	RequestConfiguration `yaml:",inline"`
	MaxRetries           int     `yaml:"max_retries" json:"maxRetries"`
	RetryInterval        float32 `yaml:"retry_interval" json:"retryInterval"`
}

type Recycle struct {
	InactivityTimeout int `yaml:"inactivity_timeout" json:"inactivityTimeout"`
}

type Commands struct {
	Start []Command `json:"start"`
	Stop  []Command `json:"stop"`
	Clean []Command `json:"clean"`
}

type Command struct {
	Command             string   `json:"command"`
	Environment         []string `yaml:"environment,omitempty" json:"environment"`
	OutputVariable      string   `yaml:"output_variable,omitempty" json:"outputVariable"`
	ContinueOnError     bool     `yaml:"continue_on_error" json:"continueOnError"`
	WorkingDir          string   `yaml:"working_dir" json:"workingDir"`
	StartHealthchecking bool     `yaml:"start_healthchecking" json:"startHealthchecking"`
	Timeout             int      `json:"timeout"`
	FireAndForget       bool     `yaml:"fire_and_forget" json:"fireAndForget"`
}

type PortConfiguration struct {
	Except []int `json:"except"`
}
