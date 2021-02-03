package services

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/phayes/freeport"
	"github.com/wufe/polo/models"
	"github.com/wufe/polo/utils"

	log "github.com/sirupsen/logrus"
)

const (
	SessionBuildResultSucceeded SessionBuildResultType = "succeeded"
	SessionBuildResultFailed    SessionBuildResultType = "failed"
)

type SessionBuildResultType string

type SessionBuildInput struct {
	Checkout string
	Service  *models.Service
}

type SessionBuildResult struct {
	Result        SessionBuildResultType
	Session       *models.Session
	FailingReason string
}

func (sessionHandler *SessionHandler) buildSession(input *SessionBuildInput) *SessionBuildResult {

	if sessionHandler.getTotalNumberOfSessions() >= sessionHandler.configuration.Global.MaxConcurrentSessions {
		return &SessionBuildResult{
			Result:        SessionBuildResultFailed,
			FailingReason: "Reached global maximum concurrent sessions",
		}
	}

	if sessionHandler.getNumberOfSessionsByService(input.Service) >= input.Service.MaxConcurrentSessions {
		return &SessionBuildResult{
			Result:        SessionBuildResultFailed,
			FailingReason: "Reached maximum concurrent sessions for this service",
		}
	}

	sessionUUID := uuid.NewString()
	log.Infof("[SESSION:%s] Building session.", sessionUUID)
	session := models.NewSession(&models.Session{
		UUID:     sessionUUID,
		Name:     input.Service.Name,
		Port:     0,
		Target:   "",
		Status:   models.SessionStatusStarting,
		Done:     make(chan struct{}),
		Service:  input.Service,
		Logs:     []models.Log{},
		CommitID: input.Checkout,
		Checkout: input.Checkout,
	})
	session.LogInfo(fmt.Sprintf("Creating session %s", session.UUID))

	freePort, err := sessionHandler.getFreePort(&input.Service.Port)
	if err != nil {
		log.Errorln("Could not get a free port", err)
		return &SessionBuildResult{
			Result:        SessionBuildResultFailed,
			FailingReason: "Could not get a free port",
		}
	}
	session.Port = freePort
	session.LogInfo(fmt.Sprintf("Found new free port: %d", session.Port))

	checkout, ok := input.Service.ObjectsToHashMap[input.Checkout]
	if !ok {
		log.Errorf("Could not find the hash of the selected checkout %s", input.Checkout)
		return &SessionBuildResult{
			Result:        SessionBuildResultFailed,
			FailingReason: fmt.Sprintf("Could not find the hash of the selected checkout %s", input.Checkout),
		}
	}
	session.CommitID = checkout
	session.LogInfo(fmt.Sprintf("Requested checkout to %s (%s)", input.Checkout, session.CommitID))

	// Check if someone else just requested the same type of session
	// looking through all open session and comparing services and checkouts
	sessionAlreadyBeingBuilt := sessionHandler.GetAliveServiceSessionByCheckout(
		checkout,
		input.Service,
	)
	if sessionAlreadyBeingBuilt != nil {
		log.Warnf(
			"Another session with the UUID %s has already being requested for checkout %s",
			sessionAlreadyBeingBuilt.UUID,
			input.Checkout,
		)
		session.LogWarn(fmt.Sprintf("Another session with the UUID %s has already being requested for checkout %s", sessionAlreadyBeingBuilt.UUID, input.Checkout))
		return &SessionBuildResult{
			Result:  SessionBuildResultSucceeded,
			Session: sessionAlreadyBeingBuilt,
		}
	}

	target := strings.ReplaceAll(input.Service.Target, "{{port}}", fmt.Sprint(freePort))
	session.Target = target
	session.LogInfo(fmt.Sprintf("Setting session target to %s", session.Target))

	session.Variables["uuid"] = session.UUID
	session.Variables["name"] = session.Name
	session.Variables["port"] = fmt.Sprint(session.Port)
	session.Variables["target"] = session.Target
	session.Variables["checkout"] = session.CommitID

	sessionHandler.sessions = append(sessionHandler.sessions, session)

	sessionStartContext, cancelSessionStart := context.WithTimeout(context.Background(), time.Second*time.Duration(session.Service.Healthcheck.RetryTimeout))
	done := make(chan struct{})

	// TODO: Persist session

	go func() {

		workingDir, err := sessionHandler.buildSessionCommitStructure(session)
		session.Folder = workingDir
		if err != nil {
			log.Errorf("Could not build session commit structure: %s", err.Error())
			cancelSessionStart()
			sessionHandler.CleanupSession(session, models.SessionStatusStartFailed)
			return
		}

		// Cleanup on context done
		go func() {
			for {
				select {
				case <-sessionStartContext.Done():
					log.Warnf("[SESSION:%s] Execution aborted", session.UUID)
					session.LogError("Execution aborted (sessionStartContext ended)")
					sessionHandler.CleanupSession(session, models.SessionStatusStartFailed)
					return
				case <-done:
					done <- struct{}{}
					return
				}
			}
		}()

		// Build the session here
		for _, command := range input.Service.Commands.Start {
			select {
			case <-sessionStartContext.Done():
				return
			default:

				if session.Status != models.SessionStatusStarting {
					cancelSessionStart()
					return
				}

				builtCommand, err := sessionHandler.buildCommand(command.Command, session)
				if err != nil {
					session.LogError(err.Error())
					log.Errorf("SESSION:%s] %s", session.UUID, err.Error())
					if !command.ContinueOnError {
						session.LogError("Halting")
						cancelSessionStart()
						return
					}
				}
				session.LogStdin(builtCommand)

				cmds := []*exec.Cmd{}
				for _, commandProg := range strings.Split(builtCommand, "|") {

					commandProg = strings.TrimSpace(commandProg)

					progAndArgs := strings.Split(commandProg, " ")

					if runtime.GOOS == "windows" {
						progAndArgs = append([]string{"cmd", "/C"}, progAndArgs...)
					}

					cmd := exec.CommandContext(sessionStartContext, progAndArgs[0], progAndArgs[1:]...)
					cmd.Env = append(
						os.Environ(),
						cmd.Env...,
					)
					cmd.Dir = sessionHandler.getWorkingDir(workingDir, command.WorkingDir)
					cmds = append(cmds, cmd)
				}

				// err = utils.ThroughCallback(utils.ExecuteCommand(cmds...))(func(line string) {
				// 	session.LogStdout(line)
				// 	log.Infof("[SESSION:%s (stdout)> ] %s", session.UUID, line)
				// 	sessionHandler.parseSessionCommandOuput(session, &command, line)
				// })

				err = utils.ExecCmds(func(line *utils.StdLine) {
					session.LogStdout(line.Line)
					log.Infof("[SESSION:%s (stdout)> ] %s", session.UUID, line.Line)
					sessionHandler.parseSessionCommandOuput(session, &command, line.Line)
				}, cmds...)

				if err != nil {
					session.LogError(err.Error())
					log.Errorf("SESSION:%s] %s", session.UUID, err.Error())
					if !command.ContinueOnError {
						session.LogError("Halting")
						cancelSessionStart()
						return
					}
				}
			}
		}

		if session.Service.Healthcheck == (models.Healthcheck{}) {
			sessionHandler.MarkSessionAsStarted(session)
			log.Infof("[SESSION:%s] Session started", session.UUID)
			done <- struct{}{}
		} else {
			// Start healthcheck routine
			// TODO: Add option to start healthchecking after N seconds (sessionStartContext should be updated accordingly)
			go func() {
				time.Sleep(5 * time.Second)
				for {
					select {
					case <-sessionStartContext.Done():
						return
					default:

						if session.Status != models.SessionStatusStarting {
							return
						}

						target, err := url.Parse(session.Target)
						if err != nil {
							session.LogError(fmt.Sprintf("Could not parse target URL: %s", err.Error()))
							log.Errorln("Could not parse target URL", err)
							cancelSessionStart()
							return
						}
						target.Path = path.Join(target.Path, input.Service.Healthcheck.URL)
						client := &http.Client{
							Timeout: 120 * time.Second,
						}
						req, err := http.NewRequest(
							input.Service.Healthcheck.Method,
							target.String(),
							nil,
						)
						req.WithContext(sessionStartContext)
						for _, header := range input.Service.Headers.Add {
							headerSegments := strings.Split(header, "=")
							req.Header.Add(headerSegments[0], headerSegments[1])
							if input.Service.Host != "" {
								req.Header.Add("Host", input.Service.Host)
							}
						}
						if err != nil {
							log.Errorln("Could not build HTTP request", req)
						}
						log.Infof("[SESSION:%s] Requesting URL %s", session.UUID, req.URL.String())
						response, err := client.Do(req)
						if err != nil {
							log.Errorf("[SESSION:%s] Could not perform HTTP request", session.UUID, err.Error())
						} else {
							if response.StatusCode == input.Service.Healthcheck.Status {
								sessionHandler.MarkSessionAsStarted(session)
								log.Infof("[SESSION:%s] Session started", session.UUID)
								done <- struct{}{}
								return
							}
						}
						log.Infof("[SESSION:%s] Session not ready yet. Retrying in %d seconds", session.UUID, session.Service.Healthcheck.RetryInterval)
						time.Sleep(time.Duration(session.Service.Healthcheck.RetryInterval) * time.Second)
					}
				}
			}()
		}

	}()

	return &SessionBuildResult{
		Result:  SessionBuildResultSucceeded,
		Session: session,
	}
}

func (sessionHandler *SessionHandler) getWorkingDir(baseDir string, commandWorkingDir string) string {
	if commandWorkingDir == "" {
		return baseDir
	}
	if strings.HasPrefix(commandWorkingDir, "./") || !strings.HasPrefix(commandWorkingDir, "/") {
		return filepath.Join(baseDir, commandWorkingDir)
	}
	return commandWorkingDir
}

func (sessionHandler *SessionHandler) parseSessionCommandOuput(session *models.Session, command *models.Command, output string) {
	session.CommandsLogs = append(session.CommandsLogs, output)
	session.Variables["last_output"] = output
	re := regexp.MustCompile(`polo\[([^\]]+?)=([^\]]+?)\]`)
	matches := re.FindAllStringSubmatch(output, -1)
	for _, variable := range matches {
		key := variable[1]
		value := variable[2]
		session.Variables[key] = value
		log.Warnf("[SESSION:%s] Setting variable %s=%s", session.UUID, key, value)
	}

	if command.OutputVariable != "" {
		session.Variables[command.OutputVariable] = output
	}
}

func (sessionHandler *SessionHandler) buildCommand(command string, session *models.Session) (string, error) {
	sessionHandler.addPortsOnDemand(command, session)
	for key, value := range session.Variables {
		command = strings.ReplaceAll(command, fmt.Sprintf("{{%s}}", key), fmt.Sprintf("%v", value))
	}
	return strings.TrimSpace(command), nil
}

func (sessionHandler *SessionHandler) addPortsOnDemand(input string, session *models.Session) (string, error) {
	re := regexp.MustCompile(`{{(port(.+?))}}`)
	matches := re.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		portVariable := match[1]
		if _, ok := session.Variables[portVariable]; !ok {
			port, err := sessionHandler.getFreePort(&session.Service.Port)
			if err != nil {
				return "", err
			}
			session.Variables[portVariable] = fmt.Sprint(port)
		}
	}
	return input, nil
}

func (sessionHandler *SessionHandler) getFreePort(portConfiguration *models.PortConfiguration) (int, error) {
	foundPort := 0
L:
	for foundPort == 0 {
		freePort, err := freeport.GetFreePort()
		if err != nil {
			return 0, err
		}
		for _, port := range portConfiguration.Except {
			if freePort == port {
				continue L
			}
		}
		foundPort = freePort
	}
	return foundPort, nil
}

func (sessionHandler *SessionHandler) getTotalNumberOfSessions() int {
	count := 0
	for _, session := range sessionHandler.sessions {
		if session.Status.IsAlive() {
			count++
		}
	}
	return count
}

func (sessionHandler *SessionHandler) getNumberOfSessionsByService(service *models.Service) int {
	count := 0
	for _, session := range sessionHandler.sessions {
		if session.Service == service && session.Status.IsAlive() {
			count++
		}
	}
	return count
}
