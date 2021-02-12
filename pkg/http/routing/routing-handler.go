package routing

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/http/proxy"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/request"
	"github.com/wufe/polo/pkg/static"
	"github.com/wufe/polo/pkg/storage"
)

type Handler struct {
	isDev              bool
	proxy              *proxy.Handler
	sessionStorage     *storage.Session
	applicationStorage *storage.Application
	request            *request.Service
	static             *static.Service
}

func NewHandler(isDev bool, proxy *proxy.Handler, sessionStorage *storage.Session, applicationStorage *storage.Application, request *request.Service, static *static.Service) *Handler {
	return &Handler{
		isDev:              isDev,
		proxy:              proxy,
		sessionStorage:     sessionStorage,
		applicationStorage: applicationStorage,
		request:            request,
		static:             static,
	}
}

func (h *Handler) RouteReverseProxyRequests() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.isDev && (strings.HasPrefix(r.URL.Path, "/_polo_") ||
			strings.HasPrefix(r.URL.Path, "/sockjs-node")) {
			h.proxy.ServeDevServer(w, r)
		} else {
			usingSmartURL := false

			var session *models.Session
			var builder proxy.Builder

			session = h.detectSession(r)

			if session == nil {
				session = h.tryGetSessionByRequestURL(r)
				if session != nil {
					usingSmartURL = true
				}
			}

			if session == nil {
				temporaryRedirect(w, "/_polo_/")
			} else {
				session.MarkAsBeingRequested()
				h.sessionStorage.Update(session)
				builder = h.buildSessionEnhancerProxy(session)

				switch session.Status {
				case models.SessionStatusStarted:
					TrackSession(w, session)
					if usingSmartURL {
						temporaryRedirect(w, "/")
					} else {
						forward := findForwardRules(r, session)
						h.serveRev(forward, builder)(w, r)
					}
					break
				case models.SessionStatusStarting:
				case models.SessionStatusDegraded:
					temporaryRedirect(w, fmt.Sprintf("/_polo_/session/%s/", session.UUID))
					break
				default:
					UntrackSession(w)
					temporaryRedirect(w, "/_polo_/")
					break
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
	return h.sessionStorage.GetByUUID(sessionUUID)
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
					session.Lock()
					serializedSession, err := json.Marshal(session)
					session.Unlock()
					if err != nil {
						serializedSession = []byte(`{}`)
					}

					serializedSession = []byte(strings.ReplaceAll(string(serializedSession), `\\`, `\\\\`))
					sessionHelper := strings.ReplaceAll(h.static.GetSessionHelperContent(), "%%currentSession%%", base64.StdEncoding.EncodeToString(serializedSession))

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

func (h *Handler) tryGetSessionByRequestURL(req *http.Request) *models.Session {
	if req.URL.Path != "" && req.URL.Path != "/" {
		appAndCheckRegexp := regexp.MustCompile(`^(/([^/]+?))?/(.+?)/?$`)
		if appAndCheck := appAndCheckRegexp.FindStringSubmatch(req.URL.Path); len(appAndCheck) == 4 {
			// Matching /<application>/<branch-seg-1>/<branch-seg-2>
			// as
			// 		application: <application>
			//		checkout: <branch-seg-1>/<branch-seg-2>
			application := appAndCheck[2]
			checkout := appAndCheck[3]
			log.Traceln("application", application)
			log.Traceln("checkout", checkout)
			foundApplication := h.applicationStorage.Get(application)
			if foundApplication == nil {

				// Matching /<branch-seg-1>/<branch-seg-2>
				// as
				// 		application: ""
				// 		<branch-seg-1>/<branch-seg-2>
				checkout = fmt.Sprintf("%s/%s", application, checkout)
				foundApplication = h.applicationStorage.Get("")
				if foundApplication == nil {
					return nil
				}

			}
			result, err := h.request.NewSession(checkout, foundApplication.Name)
			if err != nil {
				return nil
			}

			return result.Session

		}
	}
	return nil
}

func findForwardRules(req *http.Request, session *models.Session) ForwardRules {

	defaultForward, err := BuildDefaultForwardRules(session)
	if err != nil {
		panic(err)
	}

	for _, compiledPattern := range session.Application.CompiledForwardPatterns {
		if compiledPattern.Pattern.MatchString(req.URL.Path) {
			forward, err := BuildForwardRules(req, compiledPattern, session)
			if err != nil {
				return defaultForward
			}
			return forward
		}
	}
	return defaultForward
}

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
