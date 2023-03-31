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
	gitClient versioning.GitClient
	mediator  *Mediator
}

func NewSessionFilesystemWorker(gitClient versioning.GitClient, mediator *Mediator) *SessionFilesystemWorker {
	worker := &SessionFilesystemWorker{
		gitClient: gitClient,
		mediator:  mediator,
	}
	return worker
}

func (w *SessionFilesystemWorker) Start() {
	w.startAcceptingFSRequests()
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
	conf := session.GetConfiguration()
	useFolderCopy := conf.UseFolderCopy

	session.LogInfo(fmt.Sprintf("Trying to build session commit structure in folder %s", session.Application.Folder))
	checkout := sanitize.Name(session.CommitID)

	if useFolderCopy {
		return w.buildStructureCopying(session, checkout)
	}
	return w.buildStructureCloning(session, checkout)
}

func (w *SessionFilesystemWorker) buildStructureCopying(session *models.Session, checkout string) (string, error) {

	conf := session.GetConfiguration()
	applicationBaseFolder := session.Application.BaseFolder
	disableTerminalPrompt := *conf.DisableTerminalPrompt
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
	err := w.gitClient.HardReset(applicationBaseFolder, sessionCommit, versioning.WithHardResetDisableTerminalPrompt(disableTerminalPrompt))
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

func (w *SessionFilesystemWorker) buildStructureCloning(session *models.Session, checkout string) (string, error) {

	var appFolder string
	session.Application.WithRLock(func(a *models.Application) {
		appFolder = a.Folder
	})
	conf := session.GetConfiguration()
	appRemote := conf.Remote
	disableTerminalPrompt := *conf.DisableTerminalPrompt
	recurseSubmodules := *conf.RecurseSubmodules

	sessionCommitFolder := filepath.Join(appFolder, checkout)
	sessionCommit := session.CommitID

	if _, err := os.Stat(sessionCommitFolder); os.IsNotExist(err) {
		session.LogInfo(fmt.Sprintf("Cloning from remote %s into %s", appRemote, sessionCommitFolder))
		err := w.gitClient.Clone(
			appFolder,
			checkout,
			appRemote,
			versioning.WithCloneDisableTerminalPrompt(disableTerminalPrompt),
			versioning.WithCloneRecurseSubmodules(recurseSubmodules),
		)
		if err != nil {
			session.LogError(fmt.Sprintf("Error while cloning: %s", err.Error()))
			return "", err
		}
	}

	session.LogInfo("Fetching from remote")
	err := w.gitClient.FetchAll(sessionCommitFolder, versioning.WithFetchAllDisableTerminalPrompt(disableTerminalPrompt))
	if err != nil {
		session.LogError(fmt.Sprintf("Error while fetching from remote: %s", err.Error()))
		return "", err
	}

	session.LogInfo("Performing an hard reset to the selected commit")
	err = w.gitClient.HardReset(sessionCommitFolder, sessionCommit, versioning.WithHardResetDisableTerminalPrompt(disableTerminalPrompt))
	if err != nil {
		session.LogError(fmt.Sprintf("Error while performing hard reset: %s", err.Error()))
		return "", err
	}

	return sessionCommitFolder, nil
}
