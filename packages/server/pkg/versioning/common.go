package versioning

import "github.com/wufe/polo/pkg/execution"

type GitClient interface {
	Clone(baseFolder string, outputFolder string, remote string) error
	FetchAll(repoFolder string) error
	HardReset(repoFolder string, commit string) error
}

func GetGitClient(commandRunner execution.CommandRunner) GitClient {
	// Using CLI only
	return NewCLIGitClient(commandRunner)
}
