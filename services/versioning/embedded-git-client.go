package versioning

import (
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
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

func (client *EmbeddedGitClient) Clone(baseFolder string, outFolder string, remote string) error {
	_, err := git.PlainClone(filepath.Join(baseFolder, outFolder), false, &git.CloneOptions{
		URL:  remote,
		Auth: client.Auth,
	})
	return err
}

func (client *EmbeddedGitClient) FetchAll(repoFolder string) error {
	repo, err := git.PlainOpen(repoFolder)
	if err != nil {
		return err
	}
	err = repo.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"},
		Force:    true,
		Auth:     client.Auth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}
	return nil
}
