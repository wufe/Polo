package output

import "time"

type Application struct {
	Status        string                   `json:"status"`
	Configuration ApplicationConfiguration `json:"configuration"`
	Folder        string                   `json:"folder"`
	BaseFolder    string                   `json:"baseFolder"`
	BranchesMap   map[string]Branch        `json:"branchesMap"`
}

type Branch struct {
	Name    string    `json:"name"`
	Hash    string    `json:"hash"`
	Author  string    `json:"author"`
	Date    time.Time `json:"date"`
	Message string    `json:"message"`
}
