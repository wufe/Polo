package background

import (
	"context"
	"time"

	"github.com/wufe/polo/pkg/models"
)

type SessionDestroyWorker struct {
	mediator                *Mediator
	sessionCommandExecution SessionCommandExecution
}

func NewSessionDestroyWorker(mediator *Mediator, sessionCommandExecution SessionCommandExecution) *SessionDestroyWorker {
	worker := &SessionDestroyWorker{
		mediator:                mediator,
		sessionCommandExecution: sessionCommandExecution,
	}
	return worker
}

func (w *SessionDestroyWorker) Start() {
	w.startAcceptingDestroyRequests()
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

	conf := session.GetConfiguration()
	appStopCommands := conf.Commands.Stop

	session.SetStatus(models.SessionStatusStopping)
	if _, cancel, ok := session.Context.TryGet(models.SessionBuildContextKey); ok {
		cancel()
	}
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
		for _, command := range appStopCommands {
			select {
			case <-sessionStopContext.Done():
				cancelSessionStop()
				return
			default:

				err := w.sessionCommandExecution.ExecCommand(sessionStopContext, &command, session)
				if err != nil {
					session.LogError(err.Error())
					if !command.ContinueOnError {
						session.LogError("Halting")
						cancelSessionStop()
						return
					}
				}

				// builtCommand, err := buildCommand(command.Command, session)
				// if err != nil {
				// 	session.LogError(err.Error())
				// 	if !command.ContinueOnError {
				// 		session.LogError("Halting")
				// 		cancelSessionStop()
				// 		return
				// 	}
				// }

				// cmds := utils.ParseCommandContext(sessionStopContext, builtCommand)
				// for _, cmd := range cmds {
				// 	cmd.Env = append(
				// 		os.Environ(),
				// 		command.Environment...,
				// 	)
				// 	cmd.Dir = getWorkingDir(session.Folder, command.WorkingDir)
				// }
				// err = utils.ExecCmds(sessionStopContext, func(sl *utils.StdLine) {
				// 	session.LogStdout(sl.Line)
				// }, cmds...)

				// if err != nil {
				// 	session.LogError(err.Error())
				// 	if !command.ContinueOnError {
				// 		session.LogError("Halting")
				// 		cancelSessionStop()
				// 		return
				// 	}
				// }
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
