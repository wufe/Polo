package background

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/storage"
)

type SessionCleanWorker struct {
	sessionStorage          *storage.Session
	mediator                *Mediator
	sessionCommandExecution SessionCommandExecution
}

func NewSessionCleanWorker(sessionStorage *storage.Session, mediator *Mediator, sessionCommandExecution SessionCommandExecution) *SessionCleanWorker {
	worker := &SessionCleanWorker{
		sessionStorage:          sessionStorage,
		mediator:                mediator,
		sessionCommandExecution: sessionCommandExecution,
	}
	return worker
}

func (w *SessionCleanWorker) Start() {
	w.startAcceptingSessionCleanRequests()
}

func (w *SessionCleanWorker) startAcceptingSessionCleanRequests() {
	go func() {
		for {
			sessionToClean := <-w.mediator.CleanSession.Chan
			session := sessionToClean.Session

			bus := session.GetEventBus()

			killReason := session.GetKillReason()
			session.LogInfo("Cleaning up session")
			session.SetStatus(sessionToClean.Status)
			w.sessionStorage.Update(sessionToClean.Session)
			conf := session.GetConfiguration()

			// Stopping running build for current session
			if _, cancel, ok := session.Context.TryGet(models.SessionBuildContextKey); ok {
				cancel()
			}

			appCleanCommands := conf.Commands.Clean
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				sessionCleanContext, cancelSessionClean := context.WithTimeout(context.Background(), time.Second*300)

				aborted := false
				abort := func() {
					aborted = true
					cancelSessionClean()
				}

				for _, command := range appCleanCommands {
					select {
					case <-sessionCleanContext.Done():
						abort()
					default:
						bus.PublishEvent(models.SessionEventTypeCleanCommandExecution, session)

						err := w.sessionCommandExecution.ExecCommand(sessionCleanContext, &command, session)
						if err != nil {
							session.LogError(err.Error())
							if !command.ContinueOnError {
								session.LogError("Halting")
								abort()
							}
						}
					}
				}

				if aborted {
					session.LogWarn("Clean aborted")
				}

				cancelSessionClean()

				wg.Done()
			}()
			wg.Wait()

			session.LogInfo("Session cleaned up")
			appStartupRetries := conf.Startup.Retries
			appCleanOnExit := *conf.CleanOnExit

			shouldTryCleanFolders := false
			sessionGetsRecycled := false
			deleteSession := true
			sessionFailed := false

			if killReason == models.KillReasonBuildFailed || killReason == models.KillReasonHealthcheckFailed {
				maxRetries := appStartupRetries
				if maxRetries > 0 {
					retriesCount := session.GetStartupRetriesCount()
					if retriesCount < maxRetries {
						sessionGetsRecycled = true
						retriesCount++
						session.LogWarn(fmt.Sprintf("[%d/%d] Retrying session startup.", retriesCount, maxRetries))
						bus.PublishEvent(models.SessionEventTypeBuildGettingRetried, session)
						w.mediator.BuildSession.Enqueue(session.Checkout, session.Application, session, nil, false)
					} else {
						session.LogWarn("Max startup retries exceeded.")
						shouldTryCleanFolders = true
						sessionFailed = true
						deleteSession = false
					}
				} else {
					shouldTryCleanFolders = true
					sessionFailed = true
					deleteSession = false
				}
			} else {
				shouldTryCleanFolders = true
			}

			if shouldTryCleanFolders {
				if appCleanOnExit {
					bus.PublishEvent(models.SessionEventTypeFolderClean, session)
					session.LogInfo(fmt.Sprintf("Deleting session folder %s", session.Folder))
					err := os.RemoveAll(session.Folder)
					if err != nil {
						session.LogError(fmt.Sprintf("Error while removing session folder: %s", err.Error()))
					}
				}
			}

			if session.GetKillReason() == models.KillReasonStopped {
				// FEATURE: Hot swap
				// Check if the killed session should have been replaced by another session
				for _, replacement := range w.sessionStorage.GetAllAliveSessions() {
					if w.isSessionGettingReplacedBySession(session, replacement) {
						// If so, tell this session that it does not
						// replace anything anymore
						replacement.SetReplaces(nil)
						// And destroy it too
						replacement.SetKillReason(models.KillReasonStopped)
						w.mediator.DestroySession.Enqueue(replacement, nil)
						w.sessionStorage.Update(replacement)
					}
				}
			}

			if deleteSession {
				w.sessionStorage.Delete(session)
			}

			if sessionFailed {
				w.sessionStorage.AddSessionToCategory(storage.SessionCategoryFailedToStart, session)
			}

			if !sessionGetsRecycled {
				bus := session.GetEventBus()
				bus.Close()
			}

			session.Application.GetEventBus().PublishEvent(models.ApplicationEventTypeSessionCleaned, session.Application)
		}
	}()
}

func (w *SessionCleanWorker) isSessionGettingReplacedBySession(replaced *models.Session, replacement *models.Session) bool {
	for _, s := range replacement.GetReplaces() {
		if s == replaced {
			return true
		}
	}
	return false
}
