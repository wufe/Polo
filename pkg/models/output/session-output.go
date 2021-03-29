package output

import "time"

type Session struct {
	SessionStatus     `json:",inline"`
	UUID              string               `json:"uuid"`
	Name              string               `json:"name"`
	Target            string               `json:"target"`
	Port              int                  `json:"port"`
	ApplicationName   string               `json:"applicationName"`
	CommitID          string               `json:"commitID"` // The object to be checked out (branch/tag/commit id)
	CommitMessage     string               `json:"commitMessage"`
	CommitAuthorName  string               `json:"commitAuthorName"`
	CommitAuthorEmail string               `json:"commitAuthorEmail"`
	CommitDate        time.Time            `json:"commitDate"`
	CreatedAt         time.Time            `json:"createdAt"`
	Checkout          string               `json:"checkout"`
	Folder            string               `json:"folder"`
	Variables         map[string]string    `json:"variables"`
	Logs              []SessionLog         `json:"-"`
	Metrics           []Metric             `json:"metrics"`
	ReplacesSession   string               `json:"replacesSession,omitempty"`
	Configuration     SessionConfiguration `json:"configuration"`
}

type SessionConfiguration struct {
	// Application configuration
	IsDefault bool `json:"isDefault"`
	// Branch configuration
	Watch bool `json:"watch"`
}

type SessionStatus struct {
	Status     string `json:"status"`
	Age        int    `json:"age"`
	KillReason string `json:"killReason"`
	ReplacedBy string `json:"replacedBy"`
}
