package tilt

import (
	"regexp"

	tilt_models "github.com/wufe/polo/pkg/integrations/tilt/models"
)

var tiltDashboardURLRegex = regexp.MustCompile(`(?m)Tilt started on (?P<url>http[A-Za-z0-9\-._~:\/\?#\[\]@!$&'\(\)*+,;=%]+)`)

func ParseSessionCommandOutput(model tilt_models.Session, output string) tilt_models.Session {
	// Looking for tilt dashboard URL
	tiltDashboardURLRegexMatches := tiltDashboardURLRegex.FindStringSubmatch(output)
	if len(tiltDashboardURLRegexMatches) > 0 {
		dashboardURL := tiltDashboardURLRegexMatches[1]
		model.AddDashboard(dashboardURL)
	}

	return model
}
