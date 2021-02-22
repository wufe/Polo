package models

type BranchConfigurationMatch struct {
	BranchConfiguration `yaml:",inline"`
	Test                string `yaml:"test"`
}

type BranchConfiguration struct {
	Remote      string            `json:"remote"`
	Target      string            `json:"target"`
	Host        string            `json:"host"`
	Helper      Helper            `json:"helper"`
	Forwards    []Forward         `json:"forwards"`
	Headers     Headers           `json:"headers"`
	Healthcheck Healthcheck       `json:"healthCheck"`
	Startup     Startup           `json:"startup"`
	Recycle     Recycle           `json:"recycle"`
	Commands    Commands          `json:"commands"`
	Port        PortConfiguration `yaml:"port" json:"port"`
}