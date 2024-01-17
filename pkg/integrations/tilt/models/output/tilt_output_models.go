package output

type Session struct {
	Dashboards []Dashboard `json:"dashboards"`
}

// Dashboard is a struct so the ensure it will be copied by value when
// outputting the session model's output model
type Dashboard struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}
