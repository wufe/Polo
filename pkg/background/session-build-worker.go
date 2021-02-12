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
	worker.startAcceptingSessionStartRequests()

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
	return w.mediator.BuildSession.Enqueue(buildInput)
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
	session.SetStatus(models.SessionStatusStarted)
	session.SetMaxAge(session.Application.Recycle.InactivityTimeout)
	if session.GetMaxAge() > 0 {
		w.startSessionInactivityTimer(session)
	}
	w.sessionStorage.Update(session)
}

func (w *SessionBuildWorker) startSessionInactivityTimer(session *models.Session) {
	session.SetInactiveAt(time.Now().Add(time.Second * time.Duration(session.Application.Recycle.InactivityTimeout)))
	go func() {
		for {
			if session.Status != models.SessionStatusStarted {
				return
			}

			if time.Now().After(session.GetInactiveAt()) {
				w.mediator.DestroySession.Enqueue(session, nil)
				return
			}
			session.DecreaseMaxAge()
			time.Sleep(1 * time.Second)
		}
	}()
}

func (w *SessionBuildWorker) acceptSessionBuild(input *queues.SessionBuildInput) *queues.SessionBuildResult {

	aliveCount := len(w.sessionStorage.GetAllAliveSessions())
	if aliveCount >= w.global.MaxConcurrentSessions {
		return &queues.SessionBuildResult{
			Result:        queues.SessionBuildResultFailed,
			FailingReason: "Reached global maximum concurrent sessions",
		}
	}

	if w.sessionStorage.AliveByApplicationCount(input.Application) >= input.Application.MaxConcurrentSessions {
		return &queues.SessionBuildResult{
			Result:        queues.SessionBuildResultFailed,
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
		return &queues.SessionBuildResult{
			Result:  queues.SessionBuildResultSucceeded,
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

	go w.buildSession(session)

	return &queues.SessionBuildResult{
		Result:  queues.SessionBuildResultSucceeded,
		Session: session,
	}
}

func (w *SessionBuildWorker) buildSession(session *models.Session) {
	sessionStartContext, cancelSessionStart := context.WithTimeout(context.Background(), time.Second*time.Duration(session.Application.Startup.Timeout))
	defer cancelSessionStart()

	done := make(chan struct{})
	quit := make(chan struct{})
	confirm := func() {
		close(done)
	}
	abort := func() {
		close(quit)
	}

	calcBuildMetrics := models.NewMetricsForSession(session)("Build")
	err := w.prepareFolders(session)
	if err != nil {
		log.Errorf("Could not build session commit structure: %s", err.Error())
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
				log.Warnf("[SESSION:%s] Execution aborted", session.UUID)
				session.LogError("Execution aborted (sessionStartContext ended)")
				w.mediator.CleanSession.Enqueue(session, models.SessionStatusStartFailed)
				return
			case <-done:
				return
			}
		}
	}()
	healthcheckingStarted, err := w.execCommands(sessionStartContext, session, session.Application.Commands.Start)
	if err != nil {
		log.Errorf("[SESSION:%s] %s", session.UUID, err.Error())
		abort()
		return
	}

	calcBuildMetrics()
	w.sessionStorage.Update(session)

	if session.Application.Healthcheck == (models.Healthcheck{}) {
		if session.Status != models.SessionStatusStarted {
			w.mediator.StartSession.Enqueue(session)
		}
		log.Infof("[SESSION:%s] Session started", session.UUID)
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
	for _, command := range session.Application.Commands.Start {
		select {
		case <-ctx.Done():
			return healthcheckingStarted, context.Canceled
		default:

			if session.Status != models.SessionStatusStarting {
				return healthcheckingStarted, ErrWrongSessionState
			}

			err := w.execCommand(ctx, &command, session)

			if err != nil {
				session.LogError(err.Error())

				if !command.ContinueOnError {
					return healthcheckingStarted, err
				} else {
					log.Errorf("[SESSION:%s] %s", session.UUID, err.Error())
				}
			} else {
				w.sessionStorage.Update(session)
				if command.StartHealthchecking && !healthcheckingStarted && session.Application.Healthcheck != (models.Healthcheck{}) {
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
	log.Infof("[SESSION:%s (stdin)> ] %s", session.UUID, builtCommand)
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

	err = utils.ExecCmds(func(line *utils.StdLine) {
		if line.Type == utils.StdTypeOut {
			session.LogStdout(line.Line)
			log.Infof("[SESSION:%s (stdout)> ] %s", session.UUID, line.Line)
		} else {
			session.LogStderr(line.Line)
			log.Errorf("[SESSION:%s (stderr)> ] %s", session.UUID, line.Line)
		}
		parseSessionCommandOuput(session, command, line.Line)
	}, cmds...)

	return err
}
