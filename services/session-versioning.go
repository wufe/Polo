package services

import (
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/kennygrant/sanitize"
	"github.com/wufe/polo/models"
)

func (sessionHandler *SessionHandler) buildSessionCommitStructure(session *models.Session) (string, error) {
	checkout := sanitize.Name(session.Checkout)
	serviceCommitFolder := filepath.Join(session.Service.ServiceFolder, checkout)

	var repo *git.Repository

	if _, err := os.Stat(serviceCommitFolder); os.IsNotExist(err) {
		repo, err = git.PlainClone(serviceCommitFolder, false, &git.CloneOptions{
			URL: session.Service.Remote,
		})
		if err != nil {
			return "", err
		}
	} else {
		repo, err = git.PlainOpen(serviceCommitFolder)
		if err != nil {
			return "", err
		}
	}

	err := repo.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"},
		Force:    true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return "", err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	err = worktree.Reset(&git.ResetOptions{
		Mode:   git.HardReset,
		Commit: plumbing.NewHash(session.Checkout),
	})
	if err != nil {
		return "", err
	}

	return serviceCommitFolder, nil
}
