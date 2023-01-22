package models

type SharedConfiguration struct {
	Commands              Commands          `json:"commands"`
	Forwards              []Forward         `json:"forwards"`
	Headers               Headers           `json:"headers"`
	Healthcheck           Healthcheck       `json:"healthCheck"`
	Helper                Helper            `json:"helper"`
	Host                  string            `json:"host"`
	Port                  PortConfiguration `yaml:"port" json:"port"`
	Recycle               Recycle           `json:"recycle"`
	Remote                string            `json:"remote"`
	DisableTerminalPrompt *bool             `yaml:"disable_terminal_prompt" json:"disableTerminalPrompt" default:"true"`
	Startup               Startup           `json:"startup"`
	Target                string            `json:"target"`
	Warmup                Warmups           `yaml:"warmup"`
}

type Warmups struct {
	MaxRetries    int      `yaml:"max_retries"`
	RetryInterval int      `yaml:"retry_interval"`
	URLs          []Warmup `yaml:"urls"`
}

type Warmup struct {
	RequestConfiguration `yaml:",inline"`
}
