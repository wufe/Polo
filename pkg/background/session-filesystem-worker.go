package background

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/kennygrant/sanitize"
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/background/queues"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/versioning"
)

type SessionFilesystemWorker struct {
	mediator *Mediator
}

func NewSessionFilesystemWorker(mediator *Mediator) *SessionFilesystemWorker {
	worker := &SessionFilesystemWorker{
		mediator: mediator,
	}

	worker.startAcceptingFSRequests()

	return worker
}

func (w *SessionFilesystemWorker) startAcceptingFSRequests() {
	go func() {
		for {
			session := <-w.mediator.SessionFileSystem.RequestChan
			commitFolder, err := w.buildSessionCommitStructure(session)
			w.mediator.SessionFileSystem.ResponseChan <- &queues.SessionFilesystemResult{
				CommitFolder: commitFolder,
				Err:          err,
			}
		}
	}()
}

func (w *SessionFilesystemWorker) buildSessionCommitStructure(session *models.Session) (string, error) {

	log.Infof("[SESSION:%s] Trying to build session commit structure in folder %s", session.UUID, session.Application.Folder)
	session.LogInfo(fmt.Sprintf("Trying to build session commit structure in folder %s", session.Application.Folder))

	checkout := sanitize.Name(session.CommitID)

	auth, err := session.Application.GetAuth()
	if err != nil {
		log.Errorf("[SESSION:%s] Error while providing authentication %s", session.UUID, err.Error())
		session.LogError(fmt.Sprintf("Error while providing authentication: %s", err.Error()))
		return "", err
	}

	gitClient := versioning.GetGitClient(session.Application, auth)

	sessionCommitFolder := filepath.Join(session.Application.Folder, checkout)
	if _, err := os.Stat(sessionCommitFolder); os.IsNotExist(err) {
		session.LogInfo(fmt.Sprintf("Cloning from remote %s into %s", session.Application.Remote, sessionCommitFolder))
		log.Infof("[SESSION:%s] Cloning from remote %s into %s", session.UUID, session.Application.Remote, sessionCommitFolder)
		err := gitClient.Clone(session.Application.Folder, checkout, session.Application.Remote)
		if err != nil {
			session.LogError(fmt.Sprintf("Error while cloning: %s", err.Error()))
			log.Errorf("[SESSION:%s] Error while cloning: %s", session.UUID, err.Error())
			return "", err
		}
	}
	repo, err := git.PlainOpen(sessionCommitFolder)
	if err != nil {
		session.LogError(fmt.Sprintf("Error while using existing repository: %s", err.Error()))
		log.Errorf("[SESSION:%s] Error while using existing repository %s", session.UUID, err.Error())
		return "", err
	}

	session.LogInfo("Fetching from remote")
	err = gitClient.FetchAll(sessionCommitFolder)
	if err != nil {
		session.LogError(fmt.Sprintf("Error while fetching from remote: %s", err.Error()))
		log.Errorf("[SESSION:%s] Error while fetching from remote: %s", session.UUID, err.Error())
		return "", err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		session.LogError(fmt.Sprintf("Error while retrieving worktree: %s", err.Error()))
		log.Infof("[SESSION:%s] Error while retrieving worketree: %s", session.UUID, err.Error())
		return "", err
	}

	session.LogInfo("Performing an hard reset to the selected commit")
	log.Infof("[SESSION:%s] Performing an hard reset to the selected commit", session.UUID)
	err = worktree.Reset(&git.ResetOptions{
		Mode:   git.HardReset,
		Commit: plumbing.NewHash(session.CommitID),
	})
	if err != nil {
		session.LogError(fmt.Sprintf("Error while performing hard reset: %s", err.Error()))
		log.Infof("[SESSION:%s] Error while performing hard reset: %s", session.UUID, err.Error())
		return "", err
	}

	return sessionCommitFolder, nil
}
