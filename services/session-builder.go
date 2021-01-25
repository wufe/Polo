package services

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
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

	freePort, err := sessionHandler.getFreePort(&input.Service.Port)
	if err != nil {
		log.Errorln("Could not get a free port", err)
		return &SessionBuildResult{
			Result:        SessionBuildResultFailed,
			FailingReason: "Could not get a free port",
		}
	}

	session := models.NewSession(&models.Session{
		UUID:     uuid.NewString(),
		Name:     input.Service.Name,
		Port:     freePort,
		Target:   strings.ReplaceAll(input.Service.Target, "{{port}}", fmt.Sprint(freePort)),
		Status:   models.SessionStatusStarting,
		Done:     make(chan struct{}),
		Service:  input.Service,
		Logs:     []string{},
		Checkout: input.Checkout,
	})

	log.Infof("[SESSION:%s] Building session.", session.UUID)

	sessionStartContext, cancelSessionStart := context.WithTimeout(context.Background(), time.Second*time.Duration(session.Service.Healthcheck.RetryTimeout))
	done := make(chan struct{})

	go func() {

		workingDir, err := sessionHandler.buildSessionCommitStructure(session)
		session.Folder = workingDir
		if err != nil {
			log.Errorf("Could not build session commit structure: %s", err.Error())
			cancelSessionStart()
			sessionHandler.CleanupSession(session)
			return
		}

		// Cleanup on context done
		go func() {
			for {
				select {
				case <-sessionStartContext.Done():
					log.Warnf("[SESSION:%s] Execution aborted", session.UUID)
					sessionHandler.CleanupSession(session)
					return
				case <-done:
					done <- struct{}{}
					return
				}
			}
		}()

		// Start healthcheck routing
		// TODO: Add option to start healthchecking after N seconds (sessionStartContext should be updated accordingly)
		go func() {
			for {
				time.Sleep(5 * time.Second)
				select {
				case <-sessionStartContext.Done():
					return
				default:
					target, err := url.Parse(session.Target)
					if err != nil {
						log.Errorln("Could not parse target URL", err)
					}
					target.Path = path.Join(target.Path, input.Service.Healthcheck.URL)
					client := &http.Client{
						Timeout: 120,
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
						log.Errorln("Could not perform HTTP request", err)
					} else {
						if response.StatusCode == input.Service.Healthcheck.Status {
							sessionHandler.MarkSessionAsStarted(session)
							log.Infof("[SESSION:%s] Session started", session.UUID)
							done <- struct{}{}
							return
						}
					}
					log.Infof("[SESSION:%s] Session not ready yet. Retrying in 10 seconds", session.UUID)
				}
			}
		}()

		// Build the session here
		for _, command := range input.Service.Commands.Start {
			select {
			case <-sessionStartContext.Done():
				return
			default:
				commandProg := buildCommand(command.Command, session)
				progAndArgs := strings.Split(commandProg, " ")
				cmd := exec.CommandContext(sessionStartContext, progAndArgs[0], progAndArgs[1:]...)
				cmd.Env = append(
					os.Environ(),
					cmd.Env...,
				)
				cmd.Dir = workingDir

				err := utils.ThroughCallback(utils.ExecuteCommand(cmd))(func(line string) {
					log.Infof("[SESSION:%s (stdout)> ] %s", session.UUID, line)
					session.Logs = append(session.Logs, line)
				})

				if err != nil {
					log.Errorf("SESSION:%s] %s", session.UUID, err.Error())
					return
				}
			}
		}
	}()

	return &SessionBuildResult{
		Result:  SessionBuildResultSucceeded,
		Session: session,
	}
}

func buildCommand(command string, session *models.Session) string {
	replacements := make(map[string]interface{})
	replacements["uuid"] = session.UUID
	replacements["port"] = session.Port
	replacements["name"] = session.Service.Name

	for placeholder, replacement := range replacements {
		command = strings.ReplaceAll(command, fmt.Sprintf("{{%s}}", placeholder), fmt.Sprintf("%v", replacement))
	}

	return command
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
