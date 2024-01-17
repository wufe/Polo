package integrations

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	response_builder "github.com/wufe/polo/internal/rest/response-builder"
	output_models "github.com/wufe/polo/pkg/integrations/models/output"
	"github.com/wufe/polo/pkg/integrations/tilt"
	tilt_output_models "github.com/wufe/polo/pkg/integrations/tilt/models/output"
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/utils"
)

type SessionIntegrationsStatusRetriever interface {
	GetSessionIntegrationsStatus(uuid string) (*output_models.Session, error)
}

type sessionIntegrationStatusRetrieverImpl struct {
	retriever SessionIntegrationsStatusRetriever
}

func (r *sessionIntegrationStatusRetrieverImpl) GetTiltSessionsIntegrationsStatus(sessionUUID string) (*tilt_output_models.Session, error) {
	sessionIntegrationsStatus, err := r.retriever.GetSessionIntegrationsStatus(sessionUUID)
	if err != nil {
		return nil, err
	}
	return &sessionIntegrationsStatus.Tilt, nil
}

const integrationsCookieName = "PoloIntegrationSession"

type Handler struct {
	isDev  bool
	Router *httprouter.Router
	r      *response_builder.ResponseBuilder
}

func NewHandler(
	rootConfiguration *models.RootConfiguration,
	sessionIntegrationStatusRetriever SessionIntegrationsStatusRetriever,
	environment utils.Environment,
	logger logging.Logger,
) *Handler {
	router := httprouter.New()

	h := &Handler{
		isDev:  environment.IsDev(),
		Router: router,
		r:      response_builder.NewResponseBuilder(logger),
	}

	statusRetriever := &sessionIntegrationStatusRetrieverImpl{
		retriever: sessionIntegrationStatusRetriever,
	}

	tilt.InjectHandlers(router, statusRetriever, logger, integrationsCookieName)

	var catchAllHandler func(http.Handler) http.Handler = func(http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, s := h.r.NotFound()
			w.WriteHeader(s)
			w.Write(c)
		})
	}

	baseHandler := catchAllHandler(nil)

	middlewares := []func(http.Handler) http.Handler{}

	if rootConfiguration.Global.Integrations.Tilt.Enabled {
		middlewares = append(middlewares, tilt.GetDefaultCatchAllHandler(statusRetriever, logger, integrationsCookieName))
	}

	for _, middleware := range middlewares {
		baseHandler = middleware(baseHandler)
	}

	router.NotFound = baseHandler

	return h
}
