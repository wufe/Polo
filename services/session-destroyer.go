package services

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/models"
	"github.com/wufe/polo/utils"
)

func (sessionHandler *SessionHandler) DestroySession(session *models.Session) {
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
		for _, command := range session.Service.Commands.Stop {
			select {
			case <-sessionStopContext.Done():
				return
			default:
				commandProg, err := sessionHandler.buildCommand(command.Command, session)
				if err != nil {
					session.LogError(err.Error())
					log.Errorf("SESSION:%s] %s", session.UUID, err.Error())
					if !command.ContinueOnError {
						session.LogError("Halting")
						cancelSessionStop()
						return
					}
				}
				progAndArgs := strings.Split(commandProg, " ")
				cmd := exec.CommandContext(sessionStopContext, progAndArgs[0], progAndArgs[1:]...)
				cmd.Env = append(
					os.Environ(),
					cmd.Env...,
				)
				cmd.Dir = sessionHandler.getWorkingDir(session.Folder, command.WorkingDir)

				err = utils.ThroughCallback(utils.ExecuteCommand(cmd))(func(line string) {
					log.Infof("[SESSION:%s (stdout)> ] %s", session.UUID, line)
					session.LogStdout(line)
				})

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
	}()

}
