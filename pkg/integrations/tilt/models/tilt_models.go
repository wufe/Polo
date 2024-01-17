package models

import (
	"github.com/google/uuid"
	output_models "github.com/wufe/polo/pkg/integrations/tilt/models/output"
)

type Session struct {
	Dashboards []Dashboard `json:"dashboards"`
}

func NewSession(previousStoredModel *Session) *Session {
	// Not using previous stored model because dashboard data needs to be fresh
	return &Session{
		Dashboards: []Dashboard{},
	}
}

func (im *Session) AddDashboard(dashboardURL string) {
	var foundDashboard *Dashboard
	for _, d := range im.Dashboards {
		if d.URL == dashboardURL {
			foundDashboard = &d
			break
		}
	}

	if foundDashboard == nil {
		im.Dashboards = append(im.Dashboards, Dashboard{
			ID:  uuid.NewString(),
			URL: dashboardURL,
		})
	}
}

func (im *Session) ToOutput() *output_models.Session {

	copiedDashboards := make([]output_models.Dashboard, 0, len(im.Dashboards))
	for _, d := range im.Dashboards {
		copiedDashboards = append(copiedDashboards, d.ToOutput())
	}

	return &output_models.Session{
		Dashboards: copiedDashboards,
	}
}

// Dashboard is a struct so the ensure it will be copied by value when
// outputting the session model's output model
type Dashboard struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

func (d Dashboard) ToOutput() output_models.Dashboard {
	return output_models.Dashboard{
		ID:  d.ID,
		URL: d.URL,
	}
}
