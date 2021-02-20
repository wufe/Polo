package versioning

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/wufe/polo/pkg/utils"
)

type CLIGitClient struct{}

func NewCLIGitClient() GitClient {
	return &CLIGitClient{}
}

func (client *CLIGitClient) Clone(baseFolder string, outputFolder string, remote string) error {
	cmd := exec.Command("git", "clone", remote, outputFolder)
	cmd.Dir = baseFolder
	return execCommands(cmd)
}

func (client *CLIGitClient) FetchAll(repoFolder string) error {
	cmd := exec.Command("git", "fetch", "--force", "-u", "origin", "+refs/*:refs/*", "--prune")
	cmd.Dir = repoFolder
	return execCommands(cmd)
}

func (client *CLIGitClient) HardReset(repoFolder string, commit string) error {
	stash := exec.Command("git", "stash", "-u")
	stash.Dir = repoFolder

	reset := exec.Command("git", "reset", "--hard", commit)
	reset.Dir = repoFolder

	return execCommands(stash, reset)
}

func execCommands(cmds ...*exec.Cmd) error {
	for _, cmd := range cmds {
		errorLines := []string{}
		err := utils.ExecCmds(context.Background(), func(sl *utils.StdLine) {
			if sl.Type == utils.StdTypeErr {
				errorLines = append(errorLines, sl.Line)
			}
		}, cmd)
		if err != nil {
			return fmt.Errorf("%s\n%s", strings.Join(errorLines, "\n"), err.Error())
		}
	}
	return nil
}
