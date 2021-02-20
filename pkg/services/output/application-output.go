package output

import "time"

type Application struct {
	Status        string                   `json:"status"`
	Configuration ApplicationConfiguration `json:"configuration"`
	Folder        string                   `json:"folder"`
	BaseFolder    string                   `json:"baseFolder"`
	Branches      map[string]Branch        `json:"branches"`
}

type Branch struct {
	Name    string    `json:"name"`
	Hash    string    `json:"hash"`
	Author  string    `json:"author"`
	Date    time.Time `json:"date"`
	Message string    `json:"message"`
}
