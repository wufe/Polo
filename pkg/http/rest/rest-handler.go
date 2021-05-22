package rest

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/wufe/polo/pkg/http/proxy"
	"github.com/wufe/polo/pkg/http/routing"
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/models/output"
	"github.com/wufe/polo/pkg/services"
	"github.com/wufe/polo/pkg/utils"
)

type Handler struct {
	isDev  bool
	Router *httprouter.Router
	log    logging.Logger
}

func NewHandler(
	environment utils.Environment,
	static *services.StaticService,
	routing *routing.Handler,
	proxy *proxy.Handler,
	query *services.QueryService,
	request *services.RequestService,
	logger logging.Logger,
) *Handler {
	router := httprouter.New()

	h := &Handler{
		isDev:  environment.IsDev(),
		Router: router,
		log:    logger,
	}

	router.GET("/_polo_/", h.getManager(static, proxy))
	router.GET("/_polo_/session/*catchall", h.getManager(static, proxy))
	router.GET("/_polo_/api/application/", h.getApplications(query))
	router.GET("/_polo_/api/session/", h.getSessions(query))
	router.POST("/_polo_/api/session/", h.addSession(request))
	// TODO: Updated these routes to /sessions/failed/ after this PR gets merged
	// https://github.com/julienschmidt/httprouter/pull/329
	router.GET("/_polo_/api/failed/", h.getFailedSessions(query))
	router.GET("/_polo_/api/failed/:uuid", h.getFailedSession(query))
	router.GET("/_polo_/api/failed/:uuid/logs", h.getFailedSessionLogs(query))
	router.POST("/_polo_/api/failed/:uuid/ack", h.markFailedSessionAsAcknowledged(query))
	router.GET("/_polo_/api/session/:uuid", h.getSession(query))
	router.DELETE("/_polo_/api/session/:uuid", h.deleteSession(request))
	router.GET("/_polo_/api/session/:uuid/status", h.getSessionStatus(query))
	router.GET("/_polo_/api/session/:uuid/metrics", h.getSessionMetrics(query))
	router.POST("/_polo_/api/session/:uuid/track", h.trackSession(query))
	router.DELETE("/_polo_/api/session/:uuid/track", h.untrackSession())
	router.GET("/_polo_/api/session/:uuid/logs/:last_log", h.getSessionLogsAndStatus(query))
	router.GET("/_polo_/api/ping", h.ping())
	if !environment.IsDev() {
		router.GET("/_polo_/public/*filepath", h.serveStatic(static))
	}

	router.NotFound = routing.RouteReverseProxyRequests()

	return h
}

func (h *Handler) getManager(static *services.StaticService, proxy *proxy.Handler) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h.untrackSession()
		if h.isDev {
			r.URL.Path = "/_polo_/public/manager.html"
			proxy.ServeDevServer(w, r)
		} else {
			m := static.GetManager()
			w.WriteHeader(200)
			w.Write(m)
		}
	}
}

func (h *Handler) getApplications(query *services.QueryService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		content, status := h.buildResponse(ResponseObjectWithResult{
			ResponseObject{"Ok"},
			models.MapApplications(query.GetAllApplications()),
		}, 200)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write(content)
	}
}

func (h *Handler) getSessions(query *services.QueryService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		content, status := h.buildResponse(ResponseObjectWithResult{
			ResponseObject{"Ok"},
			models.MapSessions(query.GetAllAliveSessions()),
		}, 200)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write(content)
	}
}

func (h *Handler) getSession(query *services.QueryService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		session := query.GetAliveSession(uuid)

		content, status := h.okOrNotFound(session.ToOutput(), 200)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write(content)
	}
}

func (h *Handler) getSessionStatus(query *services.QueryService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		age, err := query.GetSessionStatus(uuid)

		var c []byte
		var s int

		if err != nil {
			c, s = h.notFound()
		} else {
			c, s = h.ok(age)
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(s)
		w.Write(c)
	}
}

func (h *Handler) getSessionMetrics(query *services.QueryService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		metrics, err := query.GetSessionMetrics(uuid)

		var c []byte
		var s int

		if err != nil {
			c, s = h.notFound()
		} else {
			type m struct {
				Object   string `json:"object"`
				Duration int    `json:"duration"`
			}
			ret := []m{}
			for _, metric := range metrics {
				ret = append(ret, m{
					Object:   metric.Object,
					Duration: int(metric.Duration / time.Millisecond),
				})
			}
			c, s = h.ok(ret)
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(s)
		w.Write(c)
	}
}

func (h *Handler) getSessionLogsAndStatus(query *services.QueryService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		lastLogUUID := p.ByName("last_log")
		logs, status, err := query.GetSessionLogsAndStatus(uuid, lastLogUUID)

		var c []byte
		var s int

		if err != nil {
			c, s = h.notFound()
		} else {
			c, s = h.ok(struct {
				Logs   []models.Log         `json:"logs"`
				Status models.SessionStatus `json:"status"`
			}{
				Logs:   logs,
				Status: status,
			})
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(s)
		w.Write(c)
	}
}

func (h *Handler) trackSession(query *services.QueryService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		session := query.GetAliveSession(uuid)

		var c []byte
		var s int

		if session == nil {
			c, s = h.notFound()
		} else {
			routing.TrackSession(w, session)
			c, s = h.ok(nil)
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(s)
		w.Write(c)
	}
}

func (h *Handler) untrackSession() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		routing.UntrackSession(w)

		c, s := h.ok(nil)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(s)
		w.Write(c)
	}
}

func (h *Handler) addSession(req *services.RequestService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		write := h.write(w)

		// Decoding body
		input := &struct {
			Checkout        string `json:"checkout"`
			ApplicationName string `json:"applicationName"`
		}{}
		err := json.NewDecoder(r.Body).Decode(input)
		if err != nil {
			write(h.badRequest())
			return
		}

		response, err := req.NewSession(input.Checkout, input.ApplicationName)
		if err != nil {
			if err == services.ErrApplicationNotFound {
				write(h.notFound())
				return
			}

			write(h.serverError(err.Error()))
			return
		}
		write(h.ok(response.Session.ToOutput()))
	}
}

func (h *Handler) deleteSession(req *services.RequestService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")

		write := h.write(w)

		err := req.SessionDeletion(uuid)
		if err != nil {
			switch err {
			case services.ErrSessionNotFound:
				write(h.notFound())
				return
			case services.ErrSessionIsNotAlive:
				write(h.serverError(err.Error()))
				return
			}
		}

		write(h.ok(nil))
	}
}

func (h *Handler) getFailedSessions(query *services.QueryService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		write := h.write(w)
		unacknowledged := query.GetFailedSessions()
		unacknowledgedOutput := make([]output.Session, 0, len(unacknowledged))
		for _, session := range unacknowledged {
			unacknowledgedOutput = append(unacknowledgedOutput, session.ToOutput())
		}
		acknowledged := query.GetSeenFailedSessions()
		acknowledgedOuput := make([]output.Session, 0, len(acknowledged))
		for _, session := range acknowledged {
			acknowledgedOuput = append(acknowledgedOuput, session.ToOutput())
		}
		write(h.ok(&struct {
			Unacknowledged []output.Session `json:"unacknowledged"`
			Acknowledged   []output.Session `json:"acknowledged"`
		}{unacknowledgedOutput, acknowledgedOuput}))
	}
}

func (h *Handler) getFailedSession(query *services.QueryService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		write := h.write(w)
		session, err := query.GetFailedSession(uuid)
		if err != nil {
			write(h.notFound())
		} else {
			write(h.ok(session.ToOutput()))
		}
	}
}

func (h *Handler) getFailedSessionLogs(query *services.QueryService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		write := h.write(w)
		logs, err := query.GetFailedSessionLogs(uuid)

		if err != nil {
			write(h.notFound())
		} else {
			write(h.ok(logs))
		}
	}
}

func (h *Handler) markFailedSessionAsAcknowledged(query *services.QueryService) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		write := h.write(rw)

		query.MarkFailedSessionAsSeen(uuid)

		write(h.ok(nil))
	}
}

func (h *Handler) ping() func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(200)
		w.Write([]byte("pong"))
	}
}

func (h *Handler) serveStatic(st *services.StaticService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if !h.isDev {

			fileServer := http.FileServer(st.FileSystem)

			w.Header().Set("Cache-Control", "public, max-age=604800, immutable")

			r.URL.Path = p.ByName("filepath")
			w.Header().Add("Vary", "Accept-Encoding")
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				fileServer.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()
			fileServer.ServeHTTP(&GzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
		}
	}
}

type ResponseObject struct {
	Message string `json:"message"`
}

type ResponseObjectWithResult struct {
	ResponseObject
	Result interface{} `json:"result,omitempty"`
}

type ResponseObjectWithFailingReason struct {
	ResponseObject
	Reason interface{} `json:"reason,omitempty"`
}

func (h *Handler) notFound() ([]byte, int) {
	return h.buildResponse(ResponseObjectWithFailingReason{
		ResponseObject{"Not found"},
		"Not found",
	}, 404)
}

func (h *Handler) ok(obj interface{}) ([]byte, int) {
	return h.buildResponse(ResponseObjectWithResult{
		ResponseObject{"Ok"},
		obj,
	}, 200)
}

func (h *Handler) serverError(reason interface{}) ([]byte, int) {
	return h.buildResponse(ResponseObjectWithFailingReason{
		ResponseObject{"Internal server error"},
		reason,
	}, 500)
}

func (h *Handler) badRequest() ([]byte, int) {
	return h.buildResponse(ResponseObject{"Bad request"}, 400)
}

func (h *Handler) buildResponse(response interface{}, status int) ([]byte, int) {
	responseString, err := json.Marshal(response)
	if err != nil {
		h.log.Errorln("Could not serialize response object", err)
		return []byte(`{"message": "Internal server error"}`), 500
	} else {
		return responseString, status
	}
}

func (h *Handler) okOrNotFound(obj interface{}, status int) ([]byte, int) {

	var c []byte
	var s int

	if obj != nil {
		c, s = h.ok(obj)
	} else {
		c, s = h.notFound()
	}

	return c, s
}

func (h *Handler) write(w http.ResponseWriter) func(c []byte, s int) {
	return func(c []byte, s int) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(s)
		w.Write(c)
	}
}
