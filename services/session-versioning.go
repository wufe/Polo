package services

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/kennygrant/sanitize"
	"github.com/wufe/polo/models"
)

func (sessionHandler *SessionHandler) buildSessionCommitStructure(session *models.Session) (string, error) {
	session.LogInfo(fmt.Sprintf("Trying to build session commit structure in folder %s", session.Service.ServiceFolder))

	checkout := sanitize.Name(session.Checkout)
	sessionCommitFolder := filepath.Join(session.Service.ServiceFolder, checkout)

	var repo *git.Repository

	if _, err := os.Stat(sessionCommitFolder); os.IsNotExist(err) {
		session.LogInfo(fmt.Sprintf("Cloning from remote %s into %s", session.Service.Remote, sessionCommitFolder))
		repo, err = git.PlainClone(sessionCommitFolder, false, &git.CloneOptions{
			URL: session.Service.Remote,
		})
		if err != nil {
			session.LogError(fmt.Sprintf("Error while cloning: %s", err.Error()))
			return "", err
		}
	} else {
		repo, err = git.PlainOpen(sessionCommitFolder)
		if err != nil {
			session.LogError(fmt.Sprintf("Error while using existing repository: %s", err.Error()))
			return "", err
		}
	}

	session.LogInfo("Fetching from remote")
	err := repo.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"},
		Force:    true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		session.LogError(fmt.Sprintf("Error while fetching from remote: %s", err.Error()))
		return "", err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		session.LogError(fmt.Sprintf("Error while retrieving worktree: %s", err.Error()))
		return "", err
	}

	session.LogInfo("Performing an hard reset to the selected commit")
	err = worktree.Reset(&git.ResetOptions{
		Mode:   git.HardReset,
		Commit: plumbing.NewHash(session.Checkout),
	})
	if err != nil {
		session.LogError(fmt.Sprintf("Error while performing hard reset: %s", err.Error()))
		return "", err
	}

	return sessionCommitFolder, nil
}
