package output

type ApplicationConfiguration struct {
	Name                  string            `json:"name"`
	Remote                string            `json:"remote"`
	Target                string            `json:"target"`
	Host                  string            `json:"host"`
	Fetch                 Fetch             `json:"fetch"`
	Watch                 []string          `json:"watch"`
	Helper                Helper            `json:"helper"`
	IsDefault             bool              `json:"isDefault"`
	Forwards              []Forward         `json:"forwards"`
	Headers               Headers           `json:"headers"`
	Healthcheck           Healthcheck       `json:"healthCheck"`
	Startup               Startup           `json:"startup"`
	Recycle               Recycle           `json:"recycle"`
	Commands              Commands          `json:"commands"`
	MaxConcurrentSessions int               `json:"maxConcurrentSessions"`
	Port                  PortConfiguration `json:"port"`
	UseFolderCopy         bool              `json:"useFolderCopy"`
	CleanOnExit           bool              `json:"cleanOnExit"`
}

type Fetch struct {
	Interval int `json:"interval"`
}

type Helper struct {
	Position string `json:"position"`
}

type Forward struct {
	Pattern string  `json:"pattern"`
	To      string  `json:"to"`
	Host    string  `json:"host"`
	Headers Headers `json:"headers"`
}

type Headers struct {
	Add []string `json:"add"`
	Set []string `json:"set"`
	Del []string `json:"del"`
}

type Healthcheck struct {
	Method        string `json:"method"`
	URL           string `json:"url"`
	Status        int    `json:"status"`
	MaxRetries    int    `json:"maxRetries"`
	RetryInterval int    `json:"retryInterval"`
	RetryTimeout  int    `json:"retryTimeout"`
}

type Startup struct {
	Timeout int `json:"timeout"`
	Retries int `json:"retries"`
}

type Recycle struct {
	InactivityTimeout int `json:"inactivityTimeout"`
}

type Commands struct {
	Start []Command `json:"start"`
	Stop  []Command `json:"stop"`
}

type Command struct {
	Command             string   `json:"command"`
	Environment         []string `json:"environment"`
	OutputVariable      string   `json:"outputVariable"`
	ContinueOnError     bool     `json:"continueOnError"`
	WorkingDir          string   `json:"workingDir"`
	StartHealthchecking bool     `json:"startHealthchecking"`
	Timeout             int      `json:"timeout"`
}

type PortConfiguration struct {
	Except []int `json:"except"`
}
