package output

import "time"

type Application struct {
	Status        string                   `json:"status"`
	Filename      string                   `json:"filename"`
	Configuration ApplicationConfiguration `json:"configuration"`
	Folder        string                   `json:"folder"`
	BaseFolder    string                   `json:"baseFolder"`
	BranchesMap   map[string]Branch        `json:"branchesMap"`
	TagsMap       map[string]Tag           `json:"tagsMap"`
}

type CheckoutObject struct {
	Name        string    `json:"name"`
	Hash        string    `json:"hash"`
	Author      string    `json:"author"`
	AuthorEmail string    `json:"authorEmail"`
	Date        time.Time `json:"date"`
	Message     string    `json:"message"`
}

type Branch struct {
	CheckoutObject
}

type Tag struct {
	CheckoutObject
}
