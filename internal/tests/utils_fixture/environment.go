package utils_fixture

import (
	"os"

	"github.com/wufe/polo/pkg/utils"
)

type testEnvironmentImpl struct{}

func BuildTestEnvironment() utils.Environment {
	return &testEnvironmentImpl{}
}

func (e *testEnvironmentImpl) IsTest() bool {
	return true
}

func (e *testEnvironmentImpl) IsDev() bool {
	return false
}

func (e *testEnvironmentImpl) IsDebugRace() bool {
	return false
}

func (e *testEnvironmentImpl) DevServerURL() string {
	return ""
}

func (e *testEnvironmentImpl) GetExecutableFolder() string {
	return os.Getenv("GO_CWD")
}
