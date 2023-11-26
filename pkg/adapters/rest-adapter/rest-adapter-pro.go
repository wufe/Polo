//go:build pro
// +build pro

package rest_adapter

import (
	"github.com/julienschmidt/httprouter"
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/services"
	"github.com/wufe/polo/third_party/polo-pro/http/rest"
)

func Register(router *httprouter.Router, query *services.QueryService, logger logging.Logger) {
	rest.Register(router, query, logger)
}
