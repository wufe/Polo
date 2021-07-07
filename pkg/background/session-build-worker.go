package background

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/wufe/polo/pkg/background/queues"
	"github.com/wufe/polo/pkg/http/net"
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/storage"
)

var (
	ErrWrongSessionState error = errors.New("Wrong session state")
	ErrCommandFailed     error = errors.New("Command failed")
)

type SessionBuildWorker struct {
	global                  *models.GlobalConfiguration
	applicationStorage      *storage.Application
	sessionStorage          *storage.Session
	mediator                *Mediator
	sessionBuilder          *models.SessionBuilder
	log                     logging.Logger
	sessionCommandExecution SessionCommandExecution
	portRetriever           net.PortRetriever
}

func NewSessionBuildWorker(
	globalConfiguration *models.GlobalConfiguration,
	applicationStorage *storage.Application,
	sessionStorage *storage.Session,
	mediator *Mediator,
	sessionBuilder *models.SessionBuilder,
	log logging.Logger,
	sessionCommandExecution SessionCommandExecution,
	portRetriever net.PortRetriever,
) *SessionBuildWorker {
	worker := &SessionBuildWorker{
		global:                  globalConfiguration,
		applicationStorage:      applicationStorage,
		sessionStorage:          sessionStorage,
		mediator:                mediator,
		sessionBuilder:          sessionBuilder,
		log:                     log,
		sessionCommandExecution: sessionCommandExecution,
		portRetriever:           portRetriever,
	}
	return worker
}

func (w *SessionBuildWorker) Start() {
	w.startAcceptingNewSessionRequests()
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
	return w.mediator.BuildSession.Enqueue(buildInput.Checkout, buildInput.Application, buildInput.PreviousSession, buildInput.SessionsToBeReplaced, buildInput.DetectBranchOrTag)
}

func (w *SessionBuildWorker) acceptSessionBuild(input *queues.SessionBuildInput) *queues.SessionBuildResult {

	appBus := input.Application.GetEventBus()

	conf := input.Application.GetConfiguration()
	appMaxConcurrentSessions := conf.MaxConcurrentSessions

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

	var basedOnPreviousSession bool
	var recyclingPreviousSession bool

	if input.PreviousSession != nil {
		basedOnPreviousSession = true
		killReason := input.PreviousSession.GetKillReason()
		switch killReason {
		case models.KillReasonBuildFailed, models.KillReasonHealthcheckFailed:
			recyclingPreviousSession = true
		default:
		}
	}

	var session *models.Session
	if recyclingPreviousSession {
		session = w.sessionBuilder.Build(input.PreviousSession)
		session.ResetVariables()
		session.IncStartupRetriesCount()
	} else {
		sessionUUID := uuid.NewString()
		session = w.sessionBuilder.Build(&models.Session{
			UUID:        sessionUUID,
			Port:        0,
			Status:      models.SessionStatusStarting,
			Application: input.Application,
			CommitID:    input.Checkout, // Commit ID
			Checkout:    input.Checkout, // The branch name or the commit ID
			DisplayName: input.Checkout, // The branch name or the alias
		})
	}

	// Build new alias
	sessionsNames := w.sessionStorage.GetAllSessionsNames()
	session.Alias = models.NewSessionAlias(sessionsNames)

	appBus.PublishEvent(models.ApplicationEventTypeSessionBuild, input.Application, session)

	if input.SessionsToBeReplaced != nil && len(input.SessionsToBeReplaced) > 0 {
		session.SetReplaces(input.SessionsToBeReplaced)
	}

	// Getting configuration matching this session
	conf = session.GetConfiguration()
	appPort := conf.Port

	commitID, ok := input.Application.ObjectsToHashMap[input.Checkout]
	if !ok {
		return &queues.SessionBuildResult{
			Result:        queues.SessionBuildResultFailed,
			FailingReason: fmt.Sprintf("Could not find the hash of the selected checkout %s", input.Checkout),
		}
	}

	session.LogInfo(fmt.Sprintf("Creating session %s", session.UUID))

	freePort, err := w.portRetriever.GetFreePort(appPort)
	if err != nil {
		w.log.Errorln("Could not get a free port", err)
		return &queues.SessionBuildResult{
			Result:        queues.SessionBuildResultFailed,
			FailingReason: "Could not get a free port",
		}
	}
	session.Port = freePort
	session.LogInfo(fmt.Sprintf("Found new free port: %d", session.Port))

	session.CommitID = commitID
	session.Commit = *input.Application.CommitMap[commitID]
	session.LogInfo(fmt.Sprintf("Requested checkout to %s (%s)", input.Checkout, session.CommitID))

	// Set display-name based on checkout being a commit ID or not
	// If input.detectBranchOrTag is set to true, the session's display name
	// will be set with the name of the branch or the name of the tag,
	// if given checkout name corresponds to a commitID
	_, checkoutIsTag := input.Application.TagsMap[input.Checkout]
	_, checkoutIsBranch := input.Application.BranchesMap[input.Checkout]
	object, checkoutIsObject := input.Application.HashToObjectsMap[input.Checkout]
	if input.DetectBranchOrTag && checkoutIsObject {
		// input.checkout is a commitID
		// corresponding to a branch or a tag
		if len(object.Branches) > 0 {
			session.DisplayName = object.Branches[0]
		} else if len(object.Tags) > 0 {
			session.DisplayName = object.Tags[0]
		}
	} else if !checkoutIsTag && !checkoutIsBranch {
		session.DisplayName = session.Alias
	} else {
		session.DisplayName = input.Checkout
	}

	if !basedOnPreviousSession {
		// Check if someone else just requested the same type of session
		// looking through all open session and comparing applications and checkouts
		sessionAlreadyBeingBuilt := w.sessionStorage.GetAliveApplicationSessionByCommitID(
			commitID,
			input.Application,
		)
		if sessionAlreadyBeingBuilt != nil {
			session.LogWarn(fmt.Sprintf("Another session with the UUID %s has already being requested for checkout %s", sessionAlreadyBeingBuilt.UUID, input.Checkout))
			return &queues.SessionBuildResult{
				Result:  queues.SessionBuildResultAlreadyBuilt,
				Session: sessionAlreadyBeingBuilt,
			}
		}
	}

	session.LogInfo(fmt.Sprintf("Session target is %s", session.GetTarget()))

	session.Variables["uuid"] = session.UUID
	session.Variables["name"] = session.Alias
	session.Variables["port"] = fmt.Sprint(session.Port)
	session.Variables["commit"] = session.CommitID

	w.sessionStorage.Add(session)

	go w.buildSession(session)

	return &queues.SessionBuildResult{
		Result:   queues.SessionBuildResultSucceeded,
		Session:  session,
		EventBus: session.GetEventBus(),
	}
}

func (w *SessionBuildWorker) buildSession(session *models.Session) {
	session.GetEventBus().PublishEvent(models.SessionEventTypeBuildStarted, session)
	conf := session.GetConfiguration()
	appStartupTimeout := conf.Startup.Timeout
	appHealthcheck := conf.Healthcheck

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
		session.Application.GetEventBus().PublishEvent(models.ApplicationEventTypeSessionBuildFailed, session.Application)
		close(quit)
	}

	calcBuildMetrics := models.NewMetricsForSession(session)("Build (total)")
	session.GetEventBus().PublishEvent(models.SessionEventTypePreparingFolders, session)
	err := w.prepareFolders(session)
	if err != nil {
		session.LogError(fmt.Sprintf("Could not build session commit structure: %s", err.Error()))
		session.SetKillReason(models.KillReasonBuildFailed)
		session.GetEventBus().PublishEvent(models.SessionEventTypePreparingFoldersFailed, session)
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
	healthcheckingStarted, err := w.execCommands(sessionStartContext, session, conf)
	if err != nil {
		if err == ErrWrongSessionState {
			if session.GetKillReason() == models.KillReasonNone {
				session.SetKillReason(models.KillReasonStopped)
			} else {
				session.LogTrace("Commands: it has been killed by the user, right?")
			}
		}
		session.LogError(err.Error())
		session.GetEventBus().PublishEvent(models.SessionEventTypeCommandsExecutionFailed, session)
		abort()
		return
	}

	warmup := conf.Warmup
	if len(warmup.URLs) > 0 {
		session.GetEventBus().PublishEvent(models.SessionEventTypeWarmupStarted, session)
		err := w.execWarmups(sessionStartContext, session, conf)
		if err != nil {
			if err == ErrWrongSessionState {
				if session.GetKillReason() == models.KillReasonNone {
					session.SetKillReason(models.KillReasonStopped)
				} else {
					session.LogTrace("Warmup: it has been killed by the user, right?")
				}
			}
			session.LogError(err.Error())
			session.GetEventBus().PublishEvent(models.SessionEventTypeWarmupFailed, session)
			abort()
			return
		}
	}

	calcBuildMetrics()
	w.sessionStorage.Update(session)

	if appHealthcheck == (models.Healthcheck{}) {
		if session.Status != models.SessionStatusStarted {
			w.mediator.StartSession.Enqueue(queues.SessionStartInput{
				Session: session,
			})
		}
		session.LogInfo("Session started")
	} else {
		if !healthcheckingStarted {
			w.mediator.HealthcheckSession.Enqueue(queues.SessionHealthcheckInput{
				Session: session,
			})
			healthcheckingStarted = true
		}
	}

	session.Application.GetEventBus().PublishEvent(models.ApplicationEventTypeSessionBuildSucceeded, session.Application)
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

func (w *SessionBuildWorker) execCommands(ctx context.Context, session *models.Session, conf models.ApplicationConfiguration) (healthcheckingStarted bool, err error) {
	session.GetEventBus().PublishEvent(models.SessionEventTypeCommandsExecutionStarted, session)
	calcCommandMetrics := models.NewMetricsForSession(session)("Startup commands")
	defer calcCommandMetrics()

	appHealthcheck := conf.Healthcheck
	commands := conf.Commands.Start

	for _, command := range commands {
		select {
		case <-ctx.Done():
			return healthcheckingStarted, context.Canceled
		default:

			status := session.GetStatus()
			// The command execution is permitted while the session is building or available
			if status != models.SessionStatusStarting && status != models.SessionStatusStarted {
				return healthcheckingStarted, ErrWrongSessionState
			}

			err := w.sessionCommandExecution.ExecCommand(ctx, &command, session)

			if err != nil {
				if !command.ContinueOnError {
					return healthcheckingStarted, err
				} else {
					session.LogError(err.Error())
				}
			} else {
				w.sessionStorage.Update(session)
				if command.StartHealthchecking && !healthcheckingStarted && appHealthcheck != (models.Healthcheck{}) {
					w.mediator.HealthcheckSession.Enqueue(queues.SessionHealthcheckInput{
						Session: session,
					})
					healthcheckingStarted = true
				}
			}
		}
	}
	return healthcheckingStarted, nil
}

func (w *SessionBuildWorker) execWarmups(ctx context.Context, session *models.Session, conf models.ApplicationConfiguration) error {
	calcWarmupMetrics := models.NewMetricsForSession(session)("Warmup")
	defer calcWarmupMetrics()

	warmups := conf.Warmup
	time.Sleep(1 * time.Second)

	for _, warmup := range warmups.URLs {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
			status := session.GetStatus()
			// The warmup is permitted while the session is building or availble
			if status != models.SessionStatusStarting && status != models.SessionStatusStarted {
				return ErrWrongSessionState
			}

			success, url, err := w.execWarmup(ctx, session, conf, warmup, warmups)
			if !success {
				if err != nil {
					session.LogError(fmt.Sprintf("Cannot execute warmup of URL %s: %s", url, err.Error()))
				} else {
					session.LogError(fmt.Sprintf("Cannot execute warmup of URL %s", url))
				}
			}
		}
	}

	return nil
}

func (w *SessionBuildWorker) execWarmup(ctx context.Context, session *models.Session, conf models.ApplicationConfiguration, warmup models.Warmup, warmups models.Warmups) (bool, string, error) {
	reqCtx := ctx

	retryCount := 0

	for {
		var client *http.Client
		cancelCtx := func() {}
		if warmup.Timeout != -1 {
			timeout := 60
			if warmup.Timeout > 0 {
				timeout = warmup.Timeout
			}
			timeCtx, cancel := context.WithTimeout(reqCtx, time.Duration(timeout)*time.Second)
			reqCtx = timeCtx
			cancelCtx = func() {
				cancel()
			}
			defer cancelCtx()
			client = &http.Client{
				Timeout: time.Duration(timeout) * time.Second,
			}
		} else {
			client = &http.Client{}
		}

		url := session.Variables.ApplyTo(warmup.URL)
		session.LogTrace(fmt.Sprintf("Requesting warmup URL %s", url))
		req, err := http.NewRequest(warmup.Method, url, nil)
		if err != nil {
			return false, url, err
		}
		req.WithContext(reqCtx)
		err = conf.Headers.ApplyTo(req)
		if err != nil {
			return false, url, err
		}
		if conf.Host != "" {
			req.Header.Add("Host", conf.Host)
			req.Host = conf.Host
		}
		response, err := client.Do(req)
		if err != nil || response.StatusCode != warmup.Status {

			retryCount++

			if err != nil {
				session.LogTrace(fmt.Sprintf("Warmup error: %s", err.Error()))
			} else {
				session.LogTrace(fmt.Sprintf("Warmup error: received status code %d, wanted %d", response.StatusCode, warmup.Status))
			}

			if retryCount >= warmups.MaxRetries {
				return false, url, fmt.Errorf("Warmup did not return successfull status code")
			}

			time.Sleep(time.Duration(warmups.RetryInterval) * time.Second)
			cancelCtx()
		} else {
			return true, url, nil
		}
	}
}
