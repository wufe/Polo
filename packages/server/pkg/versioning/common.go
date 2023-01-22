package versioning

import "github.com/wufe/polo/pkg/execution"

type GitClient interface {
	Clone(baseFolder string, outputFolder string, remote string, disableTerminalPrompt bool) error
	FetchAll(repoFolder string, disableTerminalPrompt bool) error
	HardReset(repoFolder string, commit string, disableTerminalPrompt bool) error
}

func GetGitClient(commandRunner execution.CommandRunner) GitClient {
	// Using CLI only
	return NewCLIGitClient(commandRunner)
}
