package services

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/kennygrant/sanitize"
	"github.com/wufe/polo/models"
	"github.com/wufe/polo/services/versioning"
)

func (sessionHandler *SessionHandler) buildSessionCommitStructure(session *models.Session) (string, error) {
	session.LogInfo(fmt.Sprintf("Trying to build session commit structure in folder %s", session.Service.ServiceFolder))

	checkout := sanitize.Name(session.CommitID)

	auth, err := session.Service.GetAuth()
	if err != nil {
		session.LogError(fmt.Sprintf("Error while providing authentication: %s", err.Error()))
		return "", err
	}

	gitClient := versioning.GetGitClient(session.Service, auth)

	sessionCommitFolder := filepath.Join(session.Service.ServiceFolder, checkout)
	if _, err := os.Stat(sessionCommitFolder); os.IsNotExist(err) {
		session.LogInfo(fmt.Sprintf("Cloning from remote %s into %s", session.Service.Remote, sessionCommitFolder))
		err := gitClient.Clone(session.Service.ServiceFolder, checkout, session.Service.Remote)
		if err != nil {
			session.LogError(fmt.Sprintf("Error while cloning: %s", err.Error()))
			return "", err
		}
	}
	repo, err := git.PlainOpen(sessionCommitFolder)
	if err != nil {
		session.LogError(fmt.Sprintf("Error while using existing repository: %s", err.Error()))
		return "", err
	}

	session.LogInfo("Fetching from remote")
	err = gitClient.FetchAll(sessionCommitFolder)
	if err != nil {
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
		Commit: plumbing.NewHash(session.CommitID),
	})
	if err != nil {
		session.LogError(fmt.Sprintf("Error while performing hard reset: %s", err.Error()))
		return "", err
	}

	return sessionCommitFolder, nil
}
