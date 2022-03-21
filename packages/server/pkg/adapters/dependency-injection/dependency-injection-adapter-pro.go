//go:build pro
// +build pro

package dependency_injection_adapter

import (
	"github.com/wufe/polo/pkg/logging"
	"go.uber.org/dig"

	dependency_injection "github.com/wufe/polo/third_party/polo-pro/pkg/dependency-injection"
)

func Register(di *dig.Container) {
	var logger logging.Logger

	di.Invoke(func(l logging.Logger) {
		logger = l
	})

	logger.Infoln("Pro build")
	logger.Debugln("Injecting pro features")

	dependency_injection.RegisterProvide(func(constructor interface{}) error {
		return di.Provide(constructor)
	})
}
