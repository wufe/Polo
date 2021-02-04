package services

import (
	"context"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/models"
	"github.com/wufe/polo/utils"
)

func (sessionHandler *SessionHandler) DestroySession(session *models.Session) {

	if !session.Status.IsAlive() {
		return
	}

	session.Status = models.SessionStatusStopping

	go func() {
		// TODO: Move that "300" into configuration
		sessionStopContext, cancelSessionStop := context.WithTimeout(context.Background(), time.Second*300)
		done := make(chan struct{})

		go func() {
			for {
				select {
				case <-sessionStopContext.Done():
					log.Warnf("[SESSION:%s] Destruction aborted", session.UUID)
					sessionHandler.CleanupSession(session, models.SessionStatusStopFailed)
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
				builtCommand, err := sessionHandler.buildCommand(command.Command, session)
				if err != nil {
					session.LogError(err.Error())
					log.Errorf("SESSION:%s] %s", session.UUID, err.Error())
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
					cmd.Dir = sessionHandler.getWorkingDir(session.Folder, command.WorkingDir)
				}
				err = utils.ExecCmds(func(sl *utils.StdLine) {
					log.Infof("[SESSION:%s (stdout)> ] %s", session.UUID, sl.Line)
					session.LogStdout(sl.Line)
				}, cmds...)

				if err != nil {
					session.LogError(err.Error())
					log.Errorf("SESSION:%s] %s", session.UUID, err.Error())
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
		sessionHandler.CleanupSession(session, models.SessionStatusStopped)

		cancelSessionStop()
	}()

}
