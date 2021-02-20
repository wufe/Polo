package output

import "time"

type Session struct {
	UUID              string            `json:"uuid"`
	Name              string            `json:"name"`
	Target            string            `json:"target"`
	Port              int               `json:"port"`
	ApplicationName   string            `json:"applicationName"`
	Status            string            `json:"status"`
	CommitID          string            `json:"commitID"` // The object to be checked out (branch/tag/commit id)
	CommitMessage     string            `json:"commitMessage"`
	CommitAuthorName  string            `json:"commitAuthorName"`
	CommitAuthorEmail string            `json:"commitAuthorEmail"`
	CommitDate        time.Time         `json:"commitDate"`
	Checkout          string            `json:"checkout"`
	MaxAge            int               `json:"maxAge"`
	Folder            string            `json:"folder"`
	Variables         map[string]string `json:"variables"`
	Logs              []SessionLog      `json:"logs"`
	Metrics           []Metric          `json:"metrics"`
}
