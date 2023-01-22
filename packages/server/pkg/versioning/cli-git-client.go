package versioning

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/wufe/polo/pkg/execution"
)

type CLIGitClient struct {
	commandRunner execution.CommandRunner
}

func NewCLIGitClient(commandRunner execution.CommandRunner) GitClient {
	return &CLIGitClient{
		commandRunner: commandRunner,
	}
}

func (client *CLIGitClient) Clone(baseFolder string, outputFolder string, remote string, disableTerminalPrompt bool) error {
	cmd := exec.Command("git", "clone", remote, outputFolder)
	if disableTerminalPrompt {
		cmd.Env = append(cmd.Env, "GIT_TERMINAL_PROMPT=0")
	}
	cmd.Dir = baseFolder
	return client.execCommands(cmd)
}

func (client *CLIGitClient) FetchAll(repoFolder string, disableTerminalPrompt bool) error {
	refsPath := path.Join(repoFolder, ".git", "refs", "remotes", "origin")
	if _, err := os.Stat(refsPath); !os.IsNotExist(err) {
		err := os.RemoveAll(refsPath)
		if err != nil {
			return err
		}
	}

	cmd := exec.Command("git", "fetch", "--force", "-u", "origin", "+refs/*:refs/*", "--prune")
	if disableTerminalPrompt {
		cmd.Env = append(cmd.Env, "GIT_TERMINAL_PROMPT=0")
	}
	cmd.Dir = repoFolder
	return client.execCommands(cmd)
}

func (client *CLIGitClient) HardReset(repoFolder string, commit string, disableTerminalPrompt bool) error {
	stash := exec.Command("git", "stash", "-u")
	if disableTerminalPrompt {
		stash.Env = append(stash.Env, "GIT_TERMINAL_PROMPT=0")
	}
	stash.Dir = repoFolder

	reset := exec.Command("git", "reset", "--hard", commit)
	if disableTerminalPrompt {
		reset.Env = append(reset.Env, "GIT_TERMINAL_PROMPT=0")
	}
	reset.Dir = repoFolder

	return client.execCommands(stash, reset)
}

func (client *CLIGitClient) execCommands(cmds ...*exec.Cmd) error {
	for _, cmd := range cmds {
		errorLines := []string{}

		err := client.commandRunner.ExecCmds(context.Background(), func(sl *execution.StdLine) {
			if sl.Type == execution.StdTypeErr {
				errorLines = append(errorLines, sl.Line)
			}
		}, cmd)
		if err != nil {
			return fmt.Errorf("%s\n%s", strings.Join(errorLines, "\n"), err.Error())
		}
	}
	return nil
}
