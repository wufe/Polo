package routing

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"

	"github.com/wufe/polo/pkg/http/proxy"
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/services"
	"github.com/wufe/polo/pkg/storage"
	"github.com/wufe/polo/pkg/utils"
)

type Handler struct {
	isDev              bool
	proxy              *proxy.Handler
	sessionStorage     *storage.Session
	applicationStorage *storage.Application
	query              *services.QueryService
	request            *services.RequestService
	static             *services.StaticService
	logger             logging.Logger
}

// NewHandler creates new routing handler
func NewHandler(environment utils.Environment, proxy *proxy.Handler, sessionStorage *storage.Session, applicationStorage *storage.Application, query *services.QueryService, request *services.RequestService, static *services.StaticService, logger logging.Logger) *Handler {
	return &Handler{
		isDev:              environment.IsDev(),
		proxy:              proxy,
		sessionStorage:     sessionStorage,
		applicationStorage: applicationStorage,
		query:              query,
		request:            request,
		static:             static,
		logger:             logger,
	}
}

// RouteReverseProxyRequests handles the routing to the right backend service.
// Each service is backed by a session, so we try to find the corresponding session.
//
// In dev mode this function checks if the request path is part of the frontend
// and serves the request proxying it to the webpack-dev-server.
//
// The session retrieval depends on the session tracking cookie value
// stored in the request object.
// If the session UUID stored in the cookie is valid, this router
// serves the backend service.
//
// If a request URL has a special "smart url" pattern
// (i.e. /s/<checkout>/<path>)
// the request is considered to be a redirect to a specific session
// identified by its checkout, in a specific path.
// The session tracking cookie value is thus skipped.
//
// Smart urls detection takes precedence over evaluation
// of session tracking cookie value.
func (h *Handler) RouteReverseProxyRequests() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.isDev && (strings.HasPrefix(r.URL.Path, "/_polo_") ||
			strings.HasPrefix(r.URL.Path, "/sockjs-node")) {
			h.proxy.ServeDevServer(w, r)
		} else {
			usingSmartURL := false

			var builder proxy.Builder

			// Here the smart url detection is performed
			session, path := h.tryGetSessionByRequestURL(r)
			if session != nil {
				usingSmartURL = true
			}
			if session == nil {
				// If smart url detection does not returns a session
				// we use session tracking cookie value to find
				// an existing session with that UUID
				session = h.detectSession(r)
			}
			if !strings.HasPrefix(r.URL.Path, "/_polo_") && (session == nil || !session.IsAlive()) {
				// FEATURE: Main branch serve
				// Retrieves default session
				session = h.getMainSession(r)
			}

			if session == nil {
				UntrackSession(w)
				temporaryRedirect(w, "/_polo_/")
			} else {
				builder = h.buildSessionEnhancerProxy(session)

				switch session.Status {
				case models.SessionStatusStarted:
					session.MarkAsBeingRequested()
					TrackSession(w, session)
					if usingSmartURL {
						// Redirects to the root, appending the path
						// got from the "smart url" pattern
						temporaryRedirect(w, fmt.Sprintf("/%s", path))
					} else {
						forward := h.findForwardRules(r, session)
						h.serveRev(forward, builder)(w, r)
					}
					break
				case models.SessionStatusStarting, models.SessionStatusDegraded:
					// Redirects to the session building page
					// appending the path given by the "smart url" pattern.
					// If the request was not generated by a smart url
					// the path will be empty string.
					temporaryRedirect(w, fmt.Sprintf("/_polo_/session/%s/%s", session.UUID, path))
				default:
					UntrackSession(w)
					temporaryRedirect(w, "/_polo_/")
				}

				if session.IsAlive() {
					h.sessionStorage.Update(session)
				}
			}
		}
	})
}

func (h *Handler) serveRev(forward ForwardRules, builder proxy.Builder) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w, r = forward(w, r)
		proxy := builder(getHostURL(r.URL))
		h.proxy.Serve(proxy)(w, r)
	}
}

func (h *Handler) detectSession(req *http.Request) *models.Session {
	cookie, err := req.Cookie("PoloSession")
	if err == http.ErrNoCookie {
		return nil
	}
	sessionUUID := cookie.Value
	session := h.sessionStorage.GetByUUID(sessionUUID)
	if session != nil {
		// FEATURE: Hot swap
		// If the found tracked session links to a replacement session
		// use that replacement UUID to look for the updated session.
		replaced := session.GetReplacedBy()
		// If it has been replaced
		if replaced != "" {
			// Search instead for the replacement
			session = h.sessionStorage.GetByUUID(replaced)
		}
	}
	return session
}

// getMainSession retrieves a session of the default application
// which is marked as "main"
// Used for retrieving a default session if no session is being tracked
func (h *Handler) getMainSession(req *http.Request) *models.Session {
	sessions := h.sessionStorage.GetAllAliveSessions()
	for _, s := range sessions {
		conf := s.GetConfiguration()
		// Its application is marked as "default"
		if conf.IsDefault {
			s.RLock()
			checkout := s.Checkout
			s.RUnlock()
			// Its branch is marked as "main"
			if conf.Branches.BranchIsMain(checkout, h.logger) {
				return s
			}
		}
	}
	return nil
}

func getHostURL(full *url.URL) *url.URL {
	url, _ := url.Parse(fmt.Sprintf("%s://%s", full.Scheme, full.Host))
	return url
}

func (h *Handler) buildSessionEnhancerProxy(session *models.Session) proxy.Builder {
	return func(url *url.URL) *httputil.ReverseProxy {
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ModifyResponse = func(r *http.Response) error {
			if strings.Contains(r.Header.Get("Content-Type"), "text/html") {
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					log.Printf("Error reading body: %v", err)
				}

				stringBody := string(body)

				var buffer *bytes.Buffer

				bodyIndexPattern := regexp.MustCompile(`<body([^>]*?)>`)

				if bodyIndex := bodyIndexPattern.FindStringIndex(stringBody); len(bodyIndex) > 1 {
					serializedSession, err := json.Marshal(session.ToOutput())
					if err != nil {
						serializedSession = []byte(`{}`)
					}

					serializedSession = []byte(strings.ReplaceAll(string(serializedSession), `\\`, `\\\\`))
					sessionHelper := strings.ReplaceAll(h.static.GetSessionHelperContent(), "%%currentSession%%", base64.StdEncoding.EncodeToString(serializedSession))

					conf := session.GetConfiguration()
					positionX, positionY := conf.Helper.Position.GetStyle()
					sessionHelper = strings.ReplaceAll(sessionHelper, "SESSION_HELPER_X", positionX)
					sessionHelper = strings.ReplaceAll(sessionHelper, "SESSION_HELPER_Y", positionY)

					stringBody = stringBody[:bodyIndex[1]] + sessionHelper + stringBody[bodyIndex[1]:]

					buffer = bytes.NewBufferString(stringBody)

				} else {
					buffer = bytes.NewBuffer(body)
				}

				r.Body = ioutil.NopCloser(buffer)
				r.Header["Content-Length"] = []string{fmt.Sprint(buffer.Len())}
			}
			return nil
		}
		return proxy
	}
}

func (h *Handler) tryGetSessionByRequestURL(req *http.Request) (*models.Session, string) {
	if strings.HasPrefix(req.URL.Path, "/s/") {
		if checkout, application, path, found := h.query.GetMatchingCheckout(req.URL.Path[3:]); found {
			result, err := h.request.NewSession(checkout, application)
			if err != nil {
				return nil, ""
			}
			if req.URL.RawQuery != "" {
				path = path + "?" + req.URL.RawQuery
			}
			return result.Session, path
		}
	}
	return nil, ""
}

func (h *Handler) findForwardRules(req *http.Request, session *models.Session) ForwardRules {
	conf := session.GetConfiguration()

	defaultForward, err := BuildDefaultForwardRules(&conf, session.Variables, h.logger)
	if err != nil {
		panic(err)
	}

	for _, compiledPattern := range session.Application.CompiledForwardPatterns {
		if compiledPattern.Pattern.MatchString(req.URL.Path) {
			forward, err := BuildForwardRules(req, compiledPattern, &conf, session.Variables, h.logger)
			if err != nil {
				return defaultForward
			}
			return forward
		}
	}
	return defaultForward
}

// TrackSession adds the session tracking cookie
// to the HTTP response
func TrackSession(res http.ResponseWriter, session *models.Session) {

	cookie := http.Cookie{
		Name:     "PoloSession",
		Value:    session.UUID,
		Path:     "/",
		MaxAge:   60 * 60 * 24,
		HttpOnly: true,
	}

	http.SetCookie(res, &cookie)
}

// UntrackSession removes the session tracking cookie
// from the HTTP response
func UntrackSession(res http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     "PoloSession",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}

	http.SetCookie(res, &cookie)
}

func temporaryRedirect(res http.ResponseWriter, location string) {
	res.Header().Add("Location", location)
	res.WriteHeader(307)
}
