package rest

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/http/proxy"
	"github.com/wufe/polo/pkg/http/routing"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/services"
)

type Handler struct {
	isDev  bool
	Router *httprouter.Router
}

func NewHandler(isDev bool, static *services.StaticService, routing *routing.Handler, proxy *proxy.Handler, query *services.QueryService, request *services.RequestService) *Handler {
	router := httprouter.New()

	h := &Handler{
		isDev:  isDev,
		Router: router,
	}

	router.GET("/_polo_/", h.getManager(static, proxy))
	router.GET("/_polo_/session/:uuid/", h.getManager(static, proxy))
	router.GET("/_polo_/api/application/", h.getApplications(query))
	router.GET("/_polo_/api/session/", h.getSessions(query))
	router.POST("/_polo_/api/session/", h.addSession(request))
	router.GET("/_polo_/api/session/:uuid", h.getSession(query))
	router.DELETE("/_polo_/api/session/:uuid", h.deleteSession(request))
	router.GET("/_polo_/api/session/:uuid/age", h.getSessionAge(query))
	router.GET("/_polo_/api/session/:uuid/metrics", h.getSessionMetrics(query))
	router.POST("/_polo_/api/session/:uuid/track", h.trackSession(query))
	router.DELETE("/_polo_/api/session/:uuid/track", h.untrackSession())
	router.GET("/_polo_/api/session/:uuid/logs/:last_log", h.getSessionLogsAndStatus(query))
	router.GET("/_polo_/api/ping", h.ping())
	if !isDev {
		router.GET("/_polo_/public/*filepath", h.serveStatic(static))
	}

	router.NotFound = routing.RouteReverseProxyRequests()

	return h
}

func (rest *Handler) getManager(static *services.StaticService, proxy *proxy.Handler) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if rest.isDev {
			r.URL.Path = "/_polo_/public/manager.html"
			proxy.ServeDevServer(w, r)
		} else {
			m := static.GetManager()
			w.WriteHeader(200)
			w.Write(m)
		}
	}
}

func (rest *Handler) getApplications(query *services.QueryService) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		content, status := buildResponse(ResponseObjectWithResult{
			ResponseObject{"Ok"},
			query.GetAllApplications(),
		}, 200)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write(content)
	}
}

func (rest *Handler) getSessions(query *services.QueryService) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		for _, s := range query.GetAllAliveSessions() {
			s.Lock()
		}

		content, status := buildResponse(ResponseObjectWithResult{
			ResponseObject{"Ok"},
			query.GetAllAliveSessions(),
		}, 200)

		for _, s := range query.GetAllAliveSessions() {
			s.Unlock()
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write(content)
	}
}

func (rest *Handler) getSession(query *services.QueryService) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		session := query.GetSession(uuid)

		content, status := okOrNotFound(session, 200)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write(content)
	}
}

func (rest *Handler) getSessionAge(query *services.QueryService) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		age, err := query.GetSessionAge(uuid)

		var c []byte
		var s int

		if err != nil {
			c, s = notFound()
		} else {
			c, s = ok(age)
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(s)
		w.Write(c)
	}
}

func (rest *Handler) getSessionMetrics(query *services.QueryService) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		metrics, err := query.GetSessionMetrics(uuid)

		var c []byte
		var s int

		if err != nil {
			c, s = notFound()
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
			c, s = ok(ret)
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(s)
		w.Write(c)
	}
}

func (rest *Handler) getSessionLogsAndStatus(query *services.QueryService) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		lastLogUUID := p.ByName("last_log")
		logs, status, err := query.GetSessionLogsAndStatus(uuid, lastLogUUID)

		var c []byte
		var s int

		if err != nil {
			c, s = notFound()
		} else {
			c, s = ok(struct {
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

func (rest *Handler) trackSession(query *services.QueryService) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		session := query.GetSession(uuid)

		var c []byte
		var s int

		if session == nil {
			c, s = notFound()
		} else {
			routing.TrackSession(w, session)
			c, s = ok(nil)
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(s)
		w.Write(c)
	}
}

func (rest *Handler) untrackSession() func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		routing.UntrackSession(w)

		c, s := ok(nil)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(s)
		w.Write(c)
	}
}

func (rest *Handler) addSession(req *services.RequestService) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		write := write(w)

		// Decoding body
		input := &struct {
			Checkout        string `json:"checkout"`
			ApplicationName string `json:"applicationName"`
		}{}
		err := json.NewDecoder(r.Body).Decode(input)
		if err != nil {
			write(badRequest())
			return
		}

		response, err := req.NewSession(input.Checkout, input.ApplicationName)
		if err != nil {
			if err == services.ErrApplicationNotFound {
				write(notFound())
				return
			}

			write(serverError(err.Error()))
			return
		}
		response.Session.Lock()
		write(ok(response.Session))
		response.Session.Unlock()
	}
}

func (rest *Handler) deleteSession(req *services.RequestService) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")

		write := write(w)

		err := req.SessionDeletion(uuid)
		if err != nil {
			switch err {
			case services.ErrSessionNotFound:
				write(notFound())
				return
			case services.ErrSessionIsNotAlive:
				write(serverError(err.Error()))
				return
			}
		}

		write(ok(nil))
	}
}

func (h *Handler) ping() func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(200)
		w.Write([]byte("pong"))
	}
}

func (h *Handler) serveStatic(st *services.StaticService) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if !h.isDev {

			fileServer := http.FileServer(st.FileSystem)

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

func notFound() ([]byte, int) {
	return buildResponse(ResponseObjectWithFailingReason{
		ResponseObject{"Not found"},
		"Not found",
	}, 404)
}

func ok(obj interface{}) ([]byte, int) {
	return buildResponse(ResponseObjectWithResult{
		ResponseObject{"Ok"},
		obj,
	}, 200)
}

func serverError(reason interface{}) ([]byte, int) {
	return buildResponse(ResponseObjectWithFailingReason{
		ResponseObject{"Internal server error"},
		reason,
	}, 500)
}

func badRequest() ([]byte, int) {
	return buildResponse(ResponseObject{"Bad request"}, 400)
}

func buildResponse(response interface{}, status int) ([]byte, int) {
	responseString, err := json.Marshal(response)
	if err != nil {
		log.Errorln("Could not serialize response object", err)
		return []byte(`{"message": "Internal server error"}`), 500
	} else {
		return responseString, status
	}
}

func okOrNotFound(obj interface{}, status int) ([]byte, int) {

	var c []byte
	var s int

	if obj != nil {
		c, s = ok(obj)
	} else {
		c, s = notFound()
	}

	return c, s
}

func write(w http.ResponseWriter) func(c []byte, s int) {
	return func(c []byte, s int) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(s)
		w.Write(c)
	}
}
