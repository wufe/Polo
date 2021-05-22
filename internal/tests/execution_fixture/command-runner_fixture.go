package execution_fixture

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/wufe/polo/pkg/execution"
)

type commandRunnerFixtureImpl struct{}

func NewCommandRunnerFixture() execution.CommandRunner {
	return &commandRunnerFixtureImpl{}
}

func (r *commandRunnerFixtureImpl) ExecCmds(ctx context.Context, callback func(*execution.StdLine), cmds ...*exec.Cmd) error {
	for _, cmd := range cmds {
		callback(&execution.StdLine{
			Type: execution.StdTypeOut,
			Line: fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " ")),
		})
	}
	return nil
}
