package versioning

import (
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

type EmbeddedGitClient struct {
	Auth transport.AuthMethod
}

func NewEmbeddedGitClient(auth transport.AuthMethod) GitClient {
	return &EmbeddedGitClient{
		Auth: auth,
	}
}

func (client *EmbeddedGitClient) Clone(baseFolder string, outputFolder string, remote string, opts ...GitCloneOpt) error {

	cfg := &GitCloneConfig{}

	for _, opt := range opts {
		opt(cfg)
	}

	var recurseSubmodules git.SubmoduleRescursivity
	if cfg.recurseSubmodules {
		recurseSubmodules = git.DefaultSubmoduleRecursionDepth
	} else {
		recurseSubmodules = git.NoRecurseSubmodules
	}

	_, err := git.PlainClone(filepath.Join(baseFolder, outputFolder), false, &git.CloneOptions{
		URL:               remote,
		Auth:              client.Auth,
		RecurseSubmodules: recurseSubmodules,
	})
	return err
}

func (client *EmbeddedGitClient) HardReset(repoFolder string, commit string, opts ...GitHardResetOpt) error {
	repo, err := git.PlainOpen(repoFolder)
	if err != nil {
		return err
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}
	return worktree.Reset(&git.ResetOptions{
		Mode:   git.HardReset,
		Commit: plumbing.NewHash(commit),
	})
}

func (client *EmbeddedGitClient) FetchAll(repoFolder string, opts ...GitFetchAllOpt) error {
	repo, err := git.PlainOpen(repoFolder)
	if err != nil {
		return err
	}
	err = repo.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{"refs/*:refs/*"},
		Force:    true,
		Auth:     client.Auth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}
	return nil
}
