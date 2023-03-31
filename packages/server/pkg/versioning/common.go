package versioning

import "github.com/wufe/polo/pkg/execution"

type GitClient interface {
	Clone(baseFolder string, outputFolder string, remote string, opts ...GitCloneOpt) error
	FetchAll(repoFolder string, opts ...GitFetchAllOpt) error
	HardReset(repoFolder string, commit string, opts ...GitHardResetOpt) error
}

func GetGitClient(commandRunner execution.CommandRunner) GitClient {
	// Using CLI only
	return NewCLIGitClient(commandRunner)
}

//region Clone
type GitCloneConfig struct {
	disableTerminalPrompt bool
	recurseSubmodules     bool
}

type GitCloneOpt = func(*GitCloneConfig)

func WithCloneDisableTerminalPrompt(value bool) GitCloneOpt {
	return func(gcc *GitCloneConfig) {
		gcc.disableTerminalPrompt = value
	}
}

func WithCloneRecurseSubmodules(value bool) GitCloneOpt {
	return func(gcc *GitCloneConfig) {
		gcc.recurseSubmodules = value
	}
}

//endregion

//region FetchAll
type GitFetchAllConfig struct {
	disableTerminalPrompt bool
}

type GitFetchAllOpt = func(*GitFetchAllConfig)

func WithFetchAllDisableTerminalPrompt(value bool) GitFetchAllOpt {
	return func(gfac *GitFetchAllConfig) {
		gfac.disableTerminalPrompt = value
	}
}

//endregion

//region HardReset
type GitHardResetConfig struct {
	disableTerminalPrompt bool
}

type GitHardResetOpt = func(*GitHardResetConfig)

func WithHardResetDisableTerminalPrompt(value bool) GitHardResetOpt {
	return func(ghrc *GitHardResetConfig) {
		ghrc.disableTerminalPrompt = value
	}
}

//endregion
