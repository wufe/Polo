package output

import "time"

type Application struct {
	Status        string                    `json:"status"`
	Filename      string                    `json:"filename"`
	Configuration ApplicationConfiguration  `json:"configuration"`
	Folder        string                    `json:"folder"`
	BaseFolder    string                    `json:"baseFolder"`
	BranchesMap   map[string]Branch         `json:"branchesMap"`
	TagsMap       map[string]Tag            `json:"tagsMap"`
	Notifications []ApplicationNotification `json:"notifications"`
}

type ApplicationNotification struct {
	UUID        string    `json:"uuid"`
	Type        string    `json:"type"`
	Permanent   bool      `json:"permanent"`
	Level       string    `json:"level"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
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
