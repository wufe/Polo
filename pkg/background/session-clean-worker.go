package background

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/storage"
	"github.com/wufe/polo/pkg/utils"
)

type SessionCleanWorker struct {
	sessionStorage *storage.Session
	mediator       *Mediator
}

func NewSessionCleanWorker(sessionStorage *storage.Session, mediator *Mediator) *SessionCleanWorker {
	worker := &SessionCleanWorker{
		sessionStorage: sessionStorage,
		mediator:       mediator,
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
						builtCommand, err := buildCommand(command.Command, session)
						if err != nil {
							session.LogError(err.Error())
							if !command.ContinueOnError {
								session.LogError("Halting")
								abort()
							}
						}

						cmds := utils.ParseCommandContext(sessionCleanContext, builtCommand)
						for _, cmd := range cmds {
							cmd.Env = append(
								os.Environ(),
								command.Environment...,
							)
							cmd.Dir = getWorkingDir(session.Folder, command.WorkingDir)
						}
						err = utils.ExecCmds(sessionCleanContext, func(sl *utils.StdLine) {
							session.LogStdout(sl.Line)
						}, cmds...)

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
						w.mediator.BuildSession.Enqueue(session.Checkout, session.Application, session)
					} else {
						shouldTryCleanFolders = true
						session.LogWarn("Max startup retries exceeded.")
						sessionFailed = true
						deleteSession = false
					}
				} else {
					shouldTryCleanFolders = true
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
					if replacement.GetReplaces() == session {
						// If so, tell this session that it is not a replacement anymore
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
		}
	}()
}
