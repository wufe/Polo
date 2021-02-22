package models

type BranchConfiguration struct {
	Remote        string            `json:"remote"`
	Target        string            `json:"target"`
	Host          string            `json:"host"`
	Helper        Helper            `json:"helper"`
	Forwards      []Forward         `json:"forwards"`
	Headers       Headers           `json:"headers"`
	Healthcheck   Healthcheck       `json:"healthCheck"`
	Startup       Startup           `json:"startup"`
	Recycle       Recycle           `json:"recycle"`
	Commands      Commands          `json:"commands"`
	Port          PortConfiguration `yaml:"port" json:"port"`
	UseFolderCopy bool              `yaml:"use_folder_copy" json:"useFolderCopy"`
	CleanOnExit   *bool             `yaml:"clean_on_exit" json:"cleanOnExit" default:"true"`
}
