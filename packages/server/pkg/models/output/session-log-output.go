package output

import "time"

type SessionLog struct {
	When    time.Time `json:"when"`
	UUID    string    `json:"uuid"`
	Type    string    `json:"type"`
	Message string    `json:"message"`
}
