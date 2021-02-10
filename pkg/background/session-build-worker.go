package background

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/background/pipe"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/storage"
	"github.com/wufe/polo/pkg/utils"
)

type SessionBuildWorker struct {
	global             *models.GlobalConfiguration
	applicationStorage *storage.Application
	sessionStorage     *storage.Session
	mediator           *Mediator
}

func NewSessionBuildWorker(
	globalConfiguration *models.GlobalConfiguration,
	applicationStorage *storage.Application,
	sessionStorage *storage.Session,
	mediator *Mediator,
) *SessionBuildWorker {
	worker := &SessionBuildWorker{
		global:             globalConfiguration,
		applicationStorage: applicationStorage,
		sessionStorage:     sessionStorage,
		mediator:           mediator,
	}

	worker.startAcceptingNewSessionRequests()
	worker.startAcceptingSessionStartRequests()

	return worker
}

func (w *SessionBuildWorker) startAcceptingNewSessionRequests() {
	go func() {
		for {
			// I'm trying to build my session.
			// Wait here until someone requests some work
			sessionBuildRequest := <-w.mediator.BuildSession.RequestChan

			sessionBuildResult := w.buildSession(sessionBuildRequest)

			w.mediator.BuildSession.ResponseChan <- sessionBuildResult
		}
	}()
}

func (w *SessionBuildWorker) RequestNewSession(buildInput *pipe.SessionBuildInput) *pipe.SessionBuildResult {
	return w.mediator.BuildSession.Request(buildInput)
}

func (w *SessionBuildWorker) startAcceptingSessionStartRequests() {
	go func() {
		for {
			session := <-w.mediator.StartSession.Chan
			w.MarkSessionAsStarted(session)
		}
	}()
}

func (w *SessionBuildWorker) MarkSessionAsStarted(session *models.Session) {
	session.Status = models.SessionStatusStarted
	session.MaxAge = session.Application.Recycle.InactivityTimeout
	if session.MaxAge > 0 {
		w.startSessionInactivityTimer(session)
	}
	w.sessionStorage.Update(session)
}

func (w *SessionBuildWorker) startSessionInactivityTimer(session *models.Session) {
	session.InactiveAt = time.Now().Add(time.Second * time.Duration(session.Application.Recycle.InactivityTimeout))
	go func() {
		for {
			if session.Status != models.SessionStatusStarted {
				return
			}

			if time.Now().After(session.InactiveAt) {
				w.mediator.DestroySession.Request(session)
				return
			}
			session.MaxAge--
			time.Sleep(1 * time.Second)
		}
	}()
}

func (w *SessionBuildWorker) buildSession(input *pipe.SessionBuildInput) *pipe.SessionBuildResult {

	aliveCount := len(w.sessionStorage.GetAllAliveSessions())
	if aliveCount >= w.global.MaxConcurrentSessions {
		return &pipe.SessionBuildResult{
			Result:        pipe.SessionBuildResultFailed,
			FailingReason: "Reached global maximum concurrent sessions",
		}
	}

	if w.sessionStorage.AliveByApplicationCount(input.Application) >= input.Application.MaxConcurrentSessions {
		return &pipe.SessionBuildResult{
			Result:        pipe.SessionBuildResultFailed,
			FailingReason: "Reached maximum concurrent sessions for this application",
		}
	}

	sessionUUID := uuid.NewString()
	log.Infof("[SESSION:%s] Building session.", sessionUUID)
	session := models.NewSession(&models.Session{
		UUID:        sessionUUID,
		Name:        input.Application.Name,
		Port:        0,
		Target:      "",
		Status:      models.SessionStatusStarting,
		Application: input.Application,
		CommitID:    input.Checkout,
		Checkout:    input.Checkout,
	})
	session.LogInfo(fmt.Sprintf("Creating session %s", session.UUID))

	freePort, err := getFreePort(&input.Application.Port)
	if err != nil {
		log.Errorln("Could not get a free port", err)
		return &pipe.SessionBuildResult{
			Result:        pipe.SessionBuildResultFailed,
			FailingReason: "Could not get a free port",
		}
	}
	session.Port = freePort
	session.LogInfo(fmt.Sprintf("Found new free port: %d", session.Port))

	checkout, ok := input.Application.ObjectsToHashMap[input.Checkout]
	if !ok {
		log.Errorf("Could not find the hash of the selected checkout %s", input.Checkout)
		return &pipe.SessionBuildResult{
			Result:        pipe.SessionBuildResultFailed,
			FailingReason: fmt.Sprintf("Could not find the hash of the selected checkout %s", input.Checkout),
		}
	}
	session.CommitID = checkout
	session.LogInfo(fmt.Sprintf("Requested checkout to %s (%s)", input.Checkout, session.CommitID))

	// Check if someone else just requested the same type of session
	// looking through all open session and comparing applications and checkouts
	sessionAlreadyBeingBuilt := w.sessionStorage.GetAliveApplicationSessionByCheckout(
		checkout,
		input.Application,
	)
	if sessionAlreadyBeingBuilt != nil {
		log.Warnf(
			"Another session with the UUID %s has already being requested for checkout %s",
			sessionAlreadyBeingBuilt.UUID,
			input.Checkout,
		)
		session.LogWarn(fmt.Sprintf("Another session with the UUID %s has already being requested for checkout %s", sessionAlreadyBeingBuilt.UUID, input.Checkout))
		return &pipe.SessionBuildResult{
			Result:  pipe.SessionBuildResultSucceeded,
			Session: sessionAlreadyBeingBuilt,
		}
	}

	target := strings.ReplaceAll(input.Application.Target, "{{port}}", fmt.Sprint(freePort))
	session.Target = target
	session.LogInfo(fmt.Sprintf("Setting session target to %s", session.Target))

	session.Variables["uuid"] = session.UUID
	session.Variables["name"] = session.Name
	session.Variables["port"] = fmt.Sprint(session.Port)
	session.Variables["target"] = session.Target
	session.Variables["commit"] = session.CommitID

	w.sessionStorage.Add(session)

	sessionStartContext, cancelSessionStart := context.WithTimeout(context.Background(), time.Second*time.Duration(session.Application.Startup.Timeout))
	done := make(chan struct{})

	go func() {

		fsResponse := w.mediator.SessionFileSystem.Request(session)
		workingDir := fsResponse.CommitFolder
		err := fsResponse.Err
		session.Folder = workingDir
		if err != nil {
			log.Errorf("Could not build session commit structure: %s", err.Error())
			cancelSessionStart()
			w.mediator.CleanSession.Request(&pipe.SessionCleanupInput{
				Session: session, Status: models.SessionStatusStartFailed,
			})
			return
		}

		// Cleanup on context done
		go func() {
			for {
				select {
				case <-sessionStartContext.Done():
					log.Warnf("[SESSION:%s] Execution aborted", session.UUID)
					session.LogError("Execution aborted (sessionStartContext ended)")
					w.mediator.CleanSession.Request(&pipe.SessionCleanupInput{
						Session: session, Status: models.SessionStatusStartFailed,
					})
					return
				case <-done:
					done <- struct{}{}
					return
				}
			}
		}()

		healthcheckingStarted := false

		// Build the session here
		for _, command := range input.Application.Commands.Start {
			select {
			case <-sessionStartContext.Done():
				return
			default:

				if session.Status != models.SessionStatusStarting {
					cancelSessionStart()
					return
				}

				builtCommand, err := buildCommand(command.Command, session)
				if err != nil {
					session.LogError(err.Error())
					log.Errorf("SESSION:%s] %s", session.UUID, err.Error())
					if !command.ContinueOnError {
						session.LogError("Halting")
						cancelSessionStart()
						return
					}
				}
				log.Infof("[SESSION:%s (stdin)> ] %s", session.UUID, builtCommand)
				session.LogStdin(builtCommand)

				cmds := utils.ParseCommandContext(sessionStartContext, builtCommand)
				for _, cmd := range cmds {
					cmd.Env = append(
						os.Environ(),
						command.Environment...,
					)
					cmd.Dir = getWorkingDir(workingDir, command.WorkingDir)
				}

				err = utils.ExecCmds(func(line *utils.StdLine) {
					session.LogStdout(line.Line)
					log.Infof("[SESSION:%s (stdout)> ] %s", session.UUID, line.Line)
					parseSessionCommandOuput(session, &command, line.Line)
				}, cmds...)

				if err != nil {
					session.LogError(err.Error())
					log.Errorf("SESSION:%s] %s", session.UUID, err.Error())
					if !command.ContinueOnError {
						session.LogError("Halting")
						cancelSessionStart()
						return
					}
				} else {
					w.sessionStorage.Update(session)
					if command.StartHealthchecking && !healthcheckingStarted && session.Application.Healthcheck != (models.Healthcheck{}) {
						w.mediator.HealthcheckSession.Request(session)
						healthcheckingStarted = true
					}
				}
			}
		}

		if session.Application.Healthcheck == (models.Healthcheck{}) {
			if session.Status != models.SessionStatusStarted {
				w.mediator.StartSession.Chan <- session
			}
			log.Infof("[SESSION:%s] Session started", session.UUID)
		} else {
			if !healthcheckingStarted {
				w.mediator.HealthcheckSession.Request(session)
				healthcheckingStarted = true
			}
		}

		done <- struct{}{}

	}()

	return &pipe.SessionBuildResult{
		Result:  pipe.SessionBuildResultSucceeded,
		Session: session,
	}
}
