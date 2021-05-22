package execution_fixture

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/wufe/polo/pkg/execution"
)

type commandRunnerFixtureImpl struct {
	sync.Mutex
	failingCommandsCount int
}

func NewCommandRunnerFixture() *commandRunnerFixtureImpl {
	return &commandRunnerFixtureImpl{
		Mutex:                sync.Mutex{},
		failingCommandsCount: 0,
	}
}

func (r *commandRunnerFixtureImpl) FailNextNCommands(n int) {
	r.Lock()
	defer r.Unlock()
	r.failingCommandsCount = n
}

func (r *commandRunnerFixtureImpl) ExecCmds(ctx context.Context, callback func(*execution.StdLine), cmds ...*exec.Cmd) error {
	if r.failingCommandsCount > 0 {
		r.Lock()
		r.failingCommandsCount = r.failingCommandsCount - 1
		r.Unlock()
		return errors.New("Command failed")
	}
	for _, cmd := range cmds {
		callback(&execution.StdLine{
			Type: execution.StdTypeOut,
			Line: fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " ")),
		})
	}
	return nil
}
