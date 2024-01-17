package tilt

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/julienschmidt/httprouter"

	response_builder "github.com/wufe/polo/internal/rest/response-builder"
	output_models "github.com/wufe/polo/pkg/integrations/tilt/models/output"
	"github.com/wufe/polo/pkg/logging"
)

type GetTiltSessionsIntegrationsStatus interface {
	GetTiltSessionsIntegrationsStatus(sessionUUID string) (*output_models.Session, error)
}

func InjectHandlers(
	router *httprouter.Router,
	query GetTiltSessionsIntegrationsStatus,
	logger logging.Logger,
	integrationsCookieName string,
) {
	router.GET("/tilt/:session_uuid/:dashboard_uuid", getSessionAndTrackTiltDashboard(query, logger, integrationsCookieName))
}

func GetDefaultCatchAllHandler(query GetTiltSessionsIntegrationsStatus, logger logging.Logger, integrationsCookieName string) func(h http.Handler) http.Handler {

	responseBuilder := response_builder.NewResponseBuilder(logger)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			cookie, err := r.Cookie(integrationsCookieName)
			if err == http.ErrNoCookie {
				h.ServeHTTP(w, r)
				return
			}

			integrationCookieValue := cookie.Value
			integrationCookieSegments := strings.Split(integrationCookieValue, "|")
			if len(integrationCookieSegments) != 3 {
				h.ServeHTTP(w, r)
				return
			}

			integrationName := integrationCookieSegments[0]
			if integrationName != "tilt" {
				h.ServeHTTP(w, r)
				return
			}

			sessionUUID := integrationCookieSegments[1]
			dashboardUUID := integrationCookieSegments[2]

			var c []byte
			var s int

			sessionIntegrationsStatus, err := query.GetTiltSessionsIntegrationsStatus(sessionUUID)
			if err != nil {
				c, s = responseBuilder.ServerError(err)
				w.WriteHeader(s)
				w.Write(c)
				return
			}

			if len(sessionIntegrationsStatus.Dashboards) == 0 {
				// TODO: Maybe redirect (after having answered with a courtesy page
				// (sth like "tilt dashboard not available anymore"), for a while)
				c, s = responseBuilder.ServerError(fmt.Errorf("no tilt dashboard found"))
				w.WriteHeader(s)
				w.Write(c)
				return
			}

			var foundDashboard *output_models.Dashboard
			for _, d := range sessionIntegrationsStatus.Dashboards {
				if d.ID == dashboardUUID {
					foundDashboard = &d
					break
				}
			}
			if foundDashboard == nil {
				// TODO: Maybe redirect (after having answered with a courtesy page
				// (sth like "tilt dashboard not available anymore"), for a while)
				c, s = responseBuilder.ServerError(fmt.Errorf("no tilt dashboard found"))
				w.WriteHeader(s)
				w.Write(c)
				return
			}

			u, _ := url.Parse(foundDashboard.URL)

			proxy := httputil.NewSingleHostReverseProxy(u)
			// TODO: Detect 502 and serve a service message page instead
			proxy.ServeHTTP(w, r)
		})
	}
}

func getSessionAndTrackTiltDashboard(query GetTiltSessionsIntegrationsStatus, logger logging.Logger, integrationsCookieName string) httprouter.Handle {

	responseBuilder := response_builder.NewResponseBuilder(logger)

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		sessionUUID := ps.ByName("session_uuid")
		dashboardUUID := ps.ByName("dashboard_uuid")

		var c []byte
		var s int

		_, err := query.GetTiltSessionsIntegrationsStatus(sessionUUID)
		if err != nil {
			c, s = responseBuilder.ServerError(err)
			w.WriteHeader(s)
			w.Write(c)
			return
		}

		cookie := http.Cookie{
			Name:     integrationsCookieName,
			Value:    "tilt|" + sessionUUID + "|" + dashboardUUID,
			Path:     "/",
			MaxAge:   60 * 60 * 24,
			HttpOnly: true,
		}

		http.SetCookie(w, &cookie)

		w.Header().Add("Location", "/")
		w.WriteHeader(307)
	}
}
