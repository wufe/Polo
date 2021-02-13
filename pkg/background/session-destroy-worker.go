package background

import (
	"context"
	"os"
	"time"

	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/utils"
)

type SessionDestroyWorker struct {
	mediator *Mediator
}

func NewSessionDestroyWorker(mediator *Mediator) *SessionDestroyWorker {
	worker := &SessionDestroyWorker{
		mediator: mediator,
	}

	worker.startAcceptingDestroyRequests()
	return worker
}

func (w *SessionDestroyWorker) startAcceptingDestroyRequests() {
	go func() {
		for {

			sessionDestroyInput := <-w.mediator.DestroySession.Chan
			w.DestroySession(sessionDestroyInput.Session, sessionDestroyInput.Callback)
		}
	}()
}

func (w *SessionDestroyWorker) DestroySession(session *models.Session, callback func(*models.Session)) {
	if !session.Status.IsAlive() {
		return
	}

	session.SetStatus(models.SessionStatusStopping)
	done := make(chan struct{})

	go func(done chan struct{}) {
		// TODO: Move that "300" into configuration
		sessionStopContext, cancelSessionStop := context.WithTimeout(context.Background(), time.Second*300)

		go func() {
			for {
				select {
				case <-sessionStopContext.Done():
					session.LogWarn("Destruction aborted")
					w.mediator.CleanSession.Enqueue(session, models.SessionStatusStopFailed)
					return
				case <-done:
					done <- struct{}{}
					return
				}
			}
		}()

		// Destroy the session here
		for _, command := range session.Application.Commands.Stop {
			select {
			case <-sessionStopContext.Done():
				cancelSessionStop()
				return
			default:
				builtCommand, err := buildCommand(command.Command, session)
				if err != nil {
					session.LogError(err.Error())
					if !command.ContinueOnError {
						session.LogError("Halting")
						cancelSessionStop()
						return
					}
				}

				cmds := utils.ParseCommandContext(sessionStopContext, builtCommand)
				for _, cmd := range cmds {
					cmd.Env = append(
						os.Environ(),
						command.Environment...,
					)
					cmd.Dir = getWorkingDir(session.Folder, command.WorkingDir)
				}
				err = utils.ExecCmds(func(sl *utils.StdLine) {
					session.LogStdout(sl.Line)
				}, cmds...)

				if err != nil {
					session.LogError(err.Error())
					if !command.ContinueOnError {
						session.LogError("Halting")
						cancelSessionStop()
						return
					}
				}
			}
		}
		done <- struct{}{}

		// In the end
		w.mediator.CleanSession.Enqueue(session, models.SessionStatusStopped)

		cancelSessionStop()

		if callback != nil {
			callback(session)
		}

	}(done)
}
