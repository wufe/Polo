//go:build pro
// +build pro

package dependency_injection_adapter

import (
	"github.com/wufe/polo/pkg/logging"
	pro_di "github.com/wufe/polo/third_party/polo-pro/di"
	"go.uber.org/dig"
)

func Register(di *dig.Container) {
	var logger logging.Logger

	di.Invoke(func(l logging.Logger) {
		logger = l
	})

	logger.Infoln("Pro build")
	logger.Debugln("Injecting pro features")

	pro_di.RegisterProvide(func(constructor interface{}) error {
		return di.Provide(constructor)
	})
}
