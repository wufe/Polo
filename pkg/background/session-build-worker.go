package background

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/background/queues"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/storage"
	"github.com/wufe/polo/pkg/utils"
)

var (
	ErrWrongSessionState error = errors.New("Wrong session state")
	ErrCommandFailed     error = errors.New("Command failed")
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

	return worker
}

func (w *SessionBuildWorker) startAcceptingNewSessionRequests() {
	go func() {
		for {
			// I'm trying to build my session.
			// Wait here until someone requests some work
			sessionBuildRequest := <-w.mediator.BuildSession.RequestChan

			sessionBuildResult := w.acceptSessionBuild(sessionBuildRequest)

			w.mediator.BuildSession.ResponseChan <- sessionBuildResult
		}
	}()
}

func (w *SessionBuildWorker) RequestNewSession(buildInput *queues.SessionBuildInput) *queues.SessionBuildResult {
	return w.mediator.BuildSession.Enqueue(buildInput.Checkout, buildInput.Application, buildInput.PreviousSession)
}

func (w *SessionBuildWorker) acceptSessionBuild(input *queues.SessionBuildInput) *queues.SessionBuildResult {

	var appName string
	var appMaxConcurrentSessions int
	var appPort models.PortConfiguration
	var appTarget string

	input.Application.Configuration.WithRLock(func(ac *models.ApplicationConfiguration) {
		appName = ac.Name
		appMaxConcurrentSessions = ac.MaxConcurrentSessions
		appPort = ac.Port
		appTarget = ac.Target
	})

	aliveCount := len(w.sessionStorage.GetAllAliveSessions())
	if aliveCount >= w.global.MaxConcurrentSessions {
		return &queues.SessionBuildResult{
			Result:        queues.SessionBuildResultFailed,
			FailingReason: "Reached global maximum concurrent sessions",
		}
	}

	if w.sessionStorage.AliveByApplicationCount(input.Application) >= appMaxConcurrentSessions {
		return &queues.SessionBuildResult{
			Result:        queues.SessionBuildResultFailed,
			FailingReason: "Reached maximum concurrent sessions for this application",
		}
	}

	var session *models.Session
	if input.PreviousSession == nil {
		sessionUUID := uuid.NewString()

		session = models.NewSession(&models.Session{
			UUID:        sessionUUID,
			Name:        appName,
			Port:        0,
			Target:      "",
			Status:      models.SessionStatusStarting,
			Application: input.Application,
			CommitID:    input.Checkout,
			Checkout:    input.Checkout,
		})
	} else {
		session = models.NewSession(input.PreviousSession)
		session.ResetVariables()
		session.IncStartupRetriesCount()
	}

	session.LogInfo(fmt.Sprintf("Creating session %s", session.UUID))

	freePort, err := getFreePort(appPort)
	if err != nil {
		log.Errorln("Could not get a free port", err)
		return &queues.SessionBuildResult{
			Result:        queues.SessionBuildResultFailed,
			FailingReason: "Could not get a free port",
		}
	}
	session.Port = freePort
	session.LogInfo(fmt.Sprintf("Found new free port: %d", session.Port))

	checkout, ok := input.Application.ObjectsToHashMap[input.Checkout]
	if !ok {
		log.Errorf("Could not find the hash of the selected checkout %s", input.Checkout)
		return &queues.SessionBuildResult{
			Result:        queues.SessionBuildResultFailed,
			FailingReason: fmt.Sprintf("Could not find the hash of the selected checkout %s", input.Checkout),
		}
	}
	session.CommitID = checkout
	session.LogInfo(fmt.Sprintf("Requested checkout to %s (%s)", input.Checkout, session.CommitID))

	recyclingPreviousSession := input.PreviousSession != nil
	if !recyclingPreviousSession {
		// Check if someone else just requested the same type of session
		// looking through all open session and comparing applications and checkouts
		sessionAlreadyBeingBuilt := w.sessionStorage.GetAliveApplicationSessionByCheckout(
			checkout,
			input.Application,
		)
		if sessionAlreadyBeingBuilt != nil {
			session.LogWarn(fmt.Sprintf("Another session with the UUID %s has already being requested for checkout %s", sessionAlreadyBeingBuilt.UUID, input.Checkout))
			return &queues.SessionBuildResult{
				Result:  queues.SessionBuildResultSucceeded,
				Session: sessionAlreadyBeingBuilt,
			}
		}
	}

	target := strings.ReplaceAll(appTarget, "{{port}}", fmt.Sprint(freePort))
	session.Target = target
	session.LogInfo(fmt.Sprintf("Setting session target to %s", session.Target))

	session.Variables["uuid"] = session.UUID
	session.Variables["name"] = session.Name
	session.Variables["port"] = fmt.Sprint(session.Port)
	session.Variables["target"] = session.Target
	session.Variables["commit"] = session.CommitID

	w.sessionStorage.Add(session)

	go w.buildSession(session)

	return &queues.SessionBuildResult{
		Result:  queues.SessionBuildResultSucceeded,
		Session: session,
	}
}

func (w *SessionBuildWorker) buildSession(session *models.Session) {
	var appStartupTimeout int
	var appStartCommands []models.Command
	var appHealthcheck models.Healthcheck

	session.Application.Configuration.WithRLock(func(ac *models.ApplicationConfiguration) {
		appStartupTimeout = ac.Startup.Timeout
		appStartCommands = ac.Commands.Start
		appHealthcheck = ac.Healthcheck
	})

	sessionStartContext, cancelSessionStart := context.WithTimeout(context.Background(), time.Second*time.Duration(appStartupTimeout))
	sessionStartContext, cancelSessionStart = context.WithCancel(sessionStartContext)
	defer session.Context.
		Named(models.SessionBuildContextKey).
		With(sessionStartContext, cancelSessionStart).
		Delete()
	defer cancelSessionStart()

	done := make(chan struct{})
	quit := make(chan struct{})
	confirm := func() {
		close(done)
	}
	abort := func() {
		close(quit)
	}

	calcBuildMetrics := models.NewMetricsForSession(session)("Build (total)")
	err := w.prepareFolders(session)
	if err != nil {
		session.LogError(fmt.Sprintf("Could not build session commit structure: %s", err.Error()))
		session.SetKillReason(models.KillReasonBuildFailed)
		abort()
		w.mediator.CleanSession.Enqueue(session, models.SessionStatusStartFailed)
		return
	}
	w.sessionStorage.Update(session)

	// Cleanup on context done
	go func() {
		for {
			select {
			case <-quit:
				session.LogError("Execution aborted")
				if session.GetKillReason() == models.KillReasonNone {
					session.SetKillReason(models.KillReasonBuildFailed)
				}
				w.mediator.CleanSession.Enqueue(session, models.SessionStatusStartFailed)
				return
			case <-done:
				return
			}
		}
	}()
	healthcheckingStarted, err := w.execCommands(sessionStartContext, session, appStartCommands)
	if err != nil {
		if err == ErrWrongSessionState {
			session.SetKillReason(models.KillReasonStopped)
		}
		session.LogError(err.Error())
		abort()
		return
	}

	calcBuildMetrics()
	w.sessionStorage.Update(session)

	if appHealthcheck == (models.Healthcheck{}) {
		if session.Status != models.SessionStatusStarted {
			w.mediator.StartSession.Enqueue(session)
		}
		session.LogInfo("Session started")
	} else {
		if !healthcheckingStarted {
			w.mediator.HealthcheckSession.Enqueue(session)
			healthcheckingStarted = true
		}
	}

	confirm()
}

func (w *SessionBuildWorker) prepareFolders(session *models.Session) error {
	calcFolderPrepareMetrics := models.NewMetricsForSession(session)("Prepare folder")
	defer calcFolderPrepareMetrics()
	fsResponse := w.mediator.SessionFileSystem.Enqueue(session)
	workingDir := fsResponse.CommitFolder
	err := fsResponse.Err
	session.Folder = workingDir
	return err
}

func (w *SessionBuildWorker) execCommands(ctx context.Context, session *models.Session, commands []models.Command) (healthcheckingStarted bool, err error) {
	calcCommandMetrics := models.NewMetricsForSession(session)("Startup commands")
	defer calcCommandMetrics()

	var appHealthcheck models.Healthcheck

	session.Application.Configuration.WithRLock(func(ac *models.ApplicationConfiguration) {
		appHealthcheck = ac.Healthcheck
	})

	for _, command := range commands {
		select {
		case <-ctx.Done():
			return healthcheckingStarted, context.Canceled
		default:

			status := session.GetStatus()
			if status != models.SessionStatusStarting {
				return healthcheckingStarted, ErrWrongSessionState
			}

			err := w.execCommand(ctx, &command, session)

			if err != nil {
				if !command.ContinueOnError {
					return healthcheckingStarted, err
				} else {
					session.LogError(err.Error())
				}
			} else {
				w.sessionStorage.Update(session)
				if command.StartHealthchecking && !healthcheckingStarted && appHealthcheck != (models.Healthcheck{}) {
					w.mediator.HealthcheckSession.Enqueue(session)
					healthcheckingStarted = true
				}
			}
		}
	}
	return healthcheckingStarted, nil
}

func (w *SessionBuildWorker) execCommand(ctx context.Context, command *models.Command, session *models.Session) error {
	builtCommand, err := buildCommand(command.Command, session)
	if err != nil {
		return err
	}
	session.LogStdin(builtCommand)

	cmdCtx := ctx
	if command.Timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(command.Timeout)*time.Second)
		defer cancel()
		cmdCtx = timeoutCtx
	}
	cmds := utils.ParseCommandContext(cmdCtx, builtCommand)
	for _, cmd := range cmds {
		cmd.Env = append(
			os.Environ(),
			command.Environment...,
		)
		cmd.Dir = getWorkingDir(session.Folder, command.WorkingDir)
	}

	err = utils.ExecCmds(ctx, func(line *utils.StdLine) {
		if line.Type == utils.StdTypeOut {
			session.LogStdout(line.Line)
		} else {
			session.LogStderr(line.Line)
		}
		parseSessionCommandOuput(session, command, line.Line)
	}, cmds...)

	return err
}
