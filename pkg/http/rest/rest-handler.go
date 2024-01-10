package rest

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	response_builder "github.com/wufe/polo/internal/rest/response-builder"
	"github.com/wufe/polo/pkg/http/proxy"
	"github.com/wufe/polo/pkg/http/routing"
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/models/output"
	"github.com/wufe/polo/pkg/services"
	"github.com/wufe/polo/pkg/utils"

	rest_adapter "github.com/wufe/polo/pkg/adapters/rest-adapter"
)

type Handler struct {
	isDev  bool
	Router *httprouter.Router
	r      *response_builder.ResponseBuilder
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
		r:      response_builder.NewResponseBuilder(logger),
	}

	router.GET("/_polo_/", h.getManager(static, proxy))
	router.GET("/_polo_/session/*catchall", h.getManager(static, proxy))
	router.GET("/_polo_/api/status", h.getStatusData(query))
	router.POST("/_polo_/api/session/", h.addSession(request))
	// TODO: Updated these routes to /sessions/failed/... after this PR gets merged
	// https://github.com/julienschmidt/httprouter/pull/329
	router.GET("/_polo_/api/failed/:uuid", h.getFailedSession(query))
	router.GET("/_polo_/api/failed/:uuid/logs", h.getFailedSessionLogs(query))
	router.POST("/_polo_/api/failed/:uuid/ack", h.markFailedSessionAsAcknowledged(query))
	router.GET("/_polo_/api/session/:uuid", h.getSession(query))
	router.DELETE("/_polo_/api/session/:uuid", h.deleteSession(request))
	router.GET("/_polo_/api/session/:uuid/status", h.getSessionStatus(query))
	router.POST("/_polo_/api/session/:uuid/track", h.trackSession(query))
	router.DELETE("/_polo_/api/session/:uuid/track", h.untrackSession())
	router.GET("/_polo_/api/session/:uuid/logs/:last_log", h.getSessionLogsAndStatus(query))
	router.GET("/_polo_/api/session/:uuid/logs", h.getSessionPTY(query, logger))
	router.GET("/_polo_/api/ping", h.ping())
	if !environment.IsDev() {
		router.GET("/_polo_/public/*filepath", h.serveStatic(static))
	}

	rest_adapter.Register(router, query, logger)

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

func (h *Handler) getStatusData(query *services.QueryService) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		write := h.r.Write(rw)

		applications := models.MapApplications(query.GetAllApplications())
		sessions := models.MapSessions(query.GetAllAliveSessions())

		unacknowledged := models.MapSessions(query.GetFailedSessions())
		acknowledged := models.MapSessions(query.GetSeenFailedSessions())

		write(h.r.Ok(StatusDataResponseObject{
			Applications: applications,
			Sessions:     sessions,
			Failures: StatusDataFailuresResponseObject{
				Unacknowledged: unacknowledged,
				Acknowledged:   acknowledged,
			},
		}))
	}
}

func (h *Handler) getSession(query *services.QueryService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		session := query.GetAliveSession(uuid)

		content, status := h.r.OkOrNotFound(session.ToOutput(), 200)

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
			c, s = h.r.NotFound()
		} else {
			c, s = h.r.Ok(age)
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
			c, s = h.r.NotFound()
		} else {
			c, s = h.r.Ok(struct {
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

func (h *Handler) getSessionPTY(query *services.QueryService, logger logging.Logger) httprouter.Handle {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		tty, _, err := query.GetSessionTTY(uuid)
		defer tty.Close()

		if err != nil {
			logger.Error("Error retrieving session TTY: %s", err)
			c, s := h.r.NotFound()
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(s)
			w.Write(c)
			return
		}

		if _, err := tty.Seek(0, io.SeekStart); err != nil {
			logger.Error("Error seeking TTY output: %s\n", err)
			c, s := h.r.ServerError(err)
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(s)
			w.Write(c)
			return
		}

		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("error while upgrading the connection: %s\n", err)
			return
		}

		var wg sync.WaitGroup
		var mutex sync.Mutex

		connectionClosed := false

		lastPong := time.Now()
		connection.SetPongHandler(func(appData string) error {
			logger.Debug("ping received")
			lastPong = time.Now()
			return nil
		})
		go func() {
			for !connectionClosed {
				mutex.Lock()

				if err := connection.WriteMessage(websocket.PingMessage, []byte("keepalive")); err != nil {
					logger.Debug("failed to write ping message")
					mutex.Unlock()
					return
				}
				mutex.Unlock()

				keepalivePingTimeout := 20 * time.Second
				time.Sleep(keepalivePingTimeout / 2) // TODO: Make it configurable
				if time.Since(lastPong) > keepalivePingTimeout {
					logger.Warn("failed to get pong response")
					connectionClosed = true
					return
				}

				logger.Debug("ping sent")
			}
		}()

		wg.Add(1)
		// tty >> xterm.js
		go func() {
			defer wg.Done()

			buffer := make([]byte, 1000)
			errorsCount := 0
			for {
				if errorsCount > 10 {
					logger.Debug("too many consecutive errors: closing")
					return
				}

				n, err := tty.Read(buffer)
				if err != nil {
					logger.Debugf("error reading from tty: %s", err)
					return
				}

				if n > 0 {
					mutex.Lock()
					if err := connection.WriteMessage(websocket.BinaryMessage, buffer[:n]); err != nil {
						logger.Debugf("error writing message to websocket: %s", err)
						errorsCount++
						mutex.Unlock()
						continue
					}
					mutex.Unlock()
				}

				errorsCount = 0
			}
		}()

		// tty << xterm.js
		go func() {
			for {
				messageType, data, err := connection.ReadMessage()
				if err != nil {
					if !connectionClosed {
						logger.Debugf("error while reading message from connection: %s", err)
					}
					return
				}

				dataBuf := bytes.Trim(data, "\x00")

				// handle resizing
				if messageType == websocket.BinaryMessage {
					if dataBuf[0] == 1 {
						ttySize := utils.TTYSize{}
						resizeMessage := bytes.Trim(dataBuf[1:], " \n\r\t\x00\x01")
						if err := json.Unmarshal(resizeMessage, &ttySize); err != nil {
							logger.Errorf("failed to unmarshal received size message '%s': %s", string(resizeMessage), err)
							continue
						}

						tty.Resize(ttySize)
						continue
					}
				}
			}
		}()

		wg.Wait()
		logger.Debug("closing websocket connection")
		connectionClosed = true
	}
}

func (h *Handler) trackSession(query *services.QueryService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		session := query.GetAliveSession(uuid)

		var c []byte
		var s int

		if session == nil {
			c, s = h.r.NotFound()
		} else {
			routing.TrackSession(w, session)
			c, s = h.r.Ok(nil)
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(s)
		w.Write(c)
	}
}

func (h *Handler) untrackSession() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		routing.UntrackSession(w)

		c, s := h.r.Ok(nil)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(s)
		w.Write(c)
	}
}

func (h *Handler) addSession(req *services.RequestService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		write := h.r.Write(w)

		// Decoding body
		input := &struct {
			Checkout        string `json:"checkout"`
			ApplicationName string `json:"applicationName"`
		}{}
		err := json.NewDecoder(r.Body).Decode(input)
		if err != nil {
			write(h.r.BadRequest())
			return
		}

		response, err := req.NewSession(input.Checkout, input.ApplicationName, false)
		if err != nil {
			if err == services.ErrApplicationNotFound {
				write(h.r.NotFound())
				return
			}

			write(h.r.ServerError(err.Error()))
			return
		}
		write(h.r.Ok(response.Session.ToOutput()))
	}
}

func (h *Handler) deleteSession(req *services.RequestService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")

		write := h.r.Write(w)

		err := req.SessionDeletion(uuid)
		if err != nil {
			switch err {
			case services.ErrSessionNotFound:
				write(h.r.NotFound())
				return
			case services.ErrSessionIsNotAlive:
				write(h.r.ServerError(err.Error()))
				return
			}
		}

		write(h.r.Ok(nil))
	}
}

func (h *Handler) getFailedSession(query *services.QueryService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		write := h.r.Write(w)
		session, err := query.GetFailedSession(uuid)
		if err != nil {
			write(h.r.NotFound())
		} else {
			write(h.r.Ok(session.ToOutput()))
		}
	}
}

func (h *Handler) getFailedSessionLogs(query *services.QueryService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		write := h.r.Write(w)
		logs, err := query.GetFailedSessionLogs(uuid)

		if err != nil {
			write(h.r.NotFound())
		} else {
			write(h.r.Ok(logs))
		}
	}
}

func (h *Handler) markFailedSessionAsAcknowledged(query *services.QueryService) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		uuid := p.ByName("uuid")
		write := h.r.Write(rw)

		query.MarkFailedSessionAsSeen(uuid)

		write(h.r.Ok(nil))
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

			//fileServer := http.FileServer(st.FileSystem)
			fileServer := http.FileServer(http.FS(st.FileSystem))

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

type StatusDataResponseObject struct {
	Applications []output.Application             `json:"applications"`
	Sessions     []output.Session                 `json:"sessions"`
	Failures     StatusDataFailuresResponseObject `json:"failures"`
}

type StatusDataFailuresResponseObject struct {
	Unacknowledged []output.Session `json:"unacknowledged"`
	Acknowledged   []output.Session `json:"acknowledged"`
}
