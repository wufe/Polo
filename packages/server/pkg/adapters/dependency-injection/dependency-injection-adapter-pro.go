// +build pro

package dependency_injection_adapter

import (
	"go.uber.org/dig"

	dependency_injection "github.com/wufe/polo/third_party/polo-pro/pkg/dependency-injection"
)

func Register(di *dig.Container) {
	dependency_injection.RegisterProvide(func(constructor interface{}) error {
		return di.Provide(constructor)
	})
}
