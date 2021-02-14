package versioning

import (
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
	return execCommand(cmd)
}

func (client *CLIGitClient) FetchAll(repoFolder string) error {
	cmd := exec.Command("git", "fetch", "--force", "-u", "origin", "+refs/*:refs/*", "--prune")
	cmd.Dir = repoFolder
	return execCommand(cmd)
}

func execCommand(cmd *exec.Cmd) error {
	errorLines := []string{}
	err := utils.ExecCmds(func(sl *utils.StdLine) {
		if sl.Type == utils.StdTypeErr {
			errorLines = append(errorLines, sl.Line)
		}
	}, cmd)
	if err != nil {
		return fmt.Errorf("%s\n%s", strings.Join(errorLines, "\n"), err.Error())
	}
	return nil
}
