package background

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kennygrant/sanitize"
	"github.com/wufe/polo/pkg/background/queues"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/utils"
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
	var useFolderCopy bool
	session.Application.Configuration.WithRLock(func(ac *models.ApplicationConfiguration) {
		useFolderCopy = ac.UseFolderCopy
	})

	session.LogInfo(fmt.Sprintf("Trying to build session commit structure in folder %s", session.Application.Folder))
	checkout := sanitize.Name(session.CommitID)

	if useFolderCopy {
		return buildStructureCopying(session, checkout)
	}
	return buildStructureCloning(session, checkout)
}

func buildStructureCopying(session *models.Session, checkout string) (string, error) {
	gitClient := versioning.GetGitClient(session.Application)

	applicationBaseFolder := session.Application.BaseFolder
	sessionCommitFolder := filepath.Join(session.Application.Folder, checkout)
	sessionCommit := session.CommitID

	// If the folder exists delete it
	if _, err := os.Stat(sessionCommitFolder); err != nil {
		session.LogInfo(fmt.Sprintf("Removing folder %s", sessionCommitFolder))
		err := os.RemoveAll(sessionCommitFolder)
		if err != nil {
			session.LogError(fmt.Sprintf("Error while deleting commit folder: %s", err.Error()))
			return "", err
		}
	}

	session.LogInfo("Performing an hard reset to the selected commit")
	err := gitClient.HardReset(applicationBaseFolder, sessionCommit)
	if err != nil {
		session.LogError(fmt.Sprintf("Error while performing hard reset: %s", err.Error()))
		return "", err
	}

	// Copy directories except .git folder
	session.LogInfo(fmt.Sprintf("Copying files from %s to %s", applicationBaseFolder, sessionCommitFolder))
	err = utils.CopyDir(applicationBaseFolder, sessionCommitFolder, func(fi os.FileInfo) bool {
		return fi.Name() != ".git"
	})

	if err != nil {
		session.LogError(fmt.Sprintf("Error while copying source directory: %s", err.Error()))
		return "", err
	}

	return sessionCommitFolder, err
}

func buildStructureCloning(session *models.Session, checkout string) (string, error) {
	gitClient := versioning.GetGitClient(session.Application)

	var appBaseFolder string
	var appFolder string
	var appRemote string
	session.Application.WithRLock(func(a *models.Application) {
		appBaseFolder = a.BaseFolder
		appFolder = a.Folder
	})
	session.Application.Configuration.WithRLock(func(ac *models.ApplicationConfiguration) {
		appRemote = ac.Remote
	})

	sessionCommitFolder := filepath.Join(appFolder, checkout)
	sessionCommit := session.CommitID

	if _, err := os.Stat(sessionCommitFolder); os.IsNotExist(err) {
		session.LogInfo(fmt.Sprintf("Cloning from remote %s into %s", appRemote, sessionCommitFolder))
		err := gitClient.Clone(appFolder, checkout, appRemote)
		if err != nil {
			session.LogError(fmt.Sprintf("Error while cloning: %s", err.Error()))
			return "", err
		}
	}

	session.LogInfo("Fetching from remote")
	err := gitClient.FetchAll(sessionCommitFolder)
	if err != nil {
		session.LogError(fmt.Sprintf("Error while fetching from remote: %s", err.Error()))
		return "", err
	}

	session.LogInfo("Performing an hard reset to the selected commit")
	err = gitClient.HardReset(appBaseFolder, sessionCommit)
	if err != nil {
		session.LogError(fmt.Sprintf("Error while performing hard reset: %s", err.Error()))
		return "", err
	}

	return sessionCommitFolder, nil
}
