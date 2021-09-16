// +build !pro

package rest_adapter

import (
	"github.com/julienschmidt/httprouter"
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/services"
)

func Register(router *httprouter.Router, query *services.QueryService, logger logging.Logger) {}
