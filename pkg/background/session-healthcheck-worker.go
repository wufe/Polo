package background

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/utils"
)

type SessionHealthcheckWorker struct {
	sessions *utils.ThreadSafeSlice
	mediator *Mediator
}

func NewSessionHealthcheckWorker(
	mediator *Mediator,
) *SessionHealthcheckWorker {
	worker := &SessionHealthcheckWorker{
		sessions: &utils.ThreadSafeSlice{
			Elements: []interface{}{},
		},
		mediator: mediator,
	}

	worker.startAcceptingSessionHealthcheckingRequests()

	return worker
}

func (w *SessionHealthcheckWorker) startAcceptingSessionHealthcheckingRequests() {
	go func() {
		for {
			session := <-w.mediator.HealthcheckSession.RequestChan
			foundSession := w.sessions.Find(session)
			if foundSession == nil {
				w.startHealthchecking(session)
			} else {
				log.Errorln("ALREADY THERE")
			}
			w.mediator.HealthcheckSession.ResponseChan <- struct{}{}
		}
	}()
}

func (w *SessionHealthcheckWorker) startHealthchecking(session *models.Session) {
	w.sessions.Push(session)
	go func() {
		time.Sleep(5 * time.Second)

		conf := session.GetConfiguration()
		maxRetries := conf.Healthcheck.MaxRetries
		healthcheck := conf.Healthcheck
		headers := conf.Headers
		host := conf.Host

		retryCount := 0

		for {
			// Failed or destroyed
			if !session.GetStatus().IsAlive() {
				w.sessions.Remove(session)
				return
			}

			target, err := url.Parse(session.Target)
			if err != nil {
				session.LogError(fmt.Sprintf("Could not parse target URL: %s", err.Error()))
				log.Errorln("Could not parse target URL", err)
				w.mediator.DestroySession.Enqueue(session, nil)
				w.sessions.Remove(session)
				return
			}
			target.Path = path.Join(target.Path, healthcheck.URL)
			client := &http.Client{
				Timeout: time.Duration(healthcheck.RetryTimeout) * time.Second,
			}
			req, err := http.NewRequest(
				healthcheck.Method,
				target.String(),
				nil,
			)
			if err != nil {
				log.Errorln("Could not build HTTP request", req)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(healthcheck.RetryTimeout)*time.Second)
			req.WithContext(ctx)
			err = headers.ApplyTo(req)
			if err != nil {
				log.Errorf("Error applying headers to the request: %s", err.Error())
			}
			if host != "" {
				req.Header.Add("Host", host)
				req.Host = host
			}
			response, err := client.Do(req)
			cancel()
			if err != nil || response.StatusCode != healthcheck.Status {
				retryCount++

				if session.Status == models.SessionStatusStarted {
					session.LogWarn("Session health degraded")
					session.SetStatus(models.SessionStatusDegraded)
				}
				if retryCount >= maxRetries {

					if session.GetStatus() == models.SessionStatusStarting {
						session.SetKillReason(models.KillReasonHealthcheckFailed)
					}

					session.LogError("Session healthcheck failed. Destroying session")
					w.mediator.DestroySession.Enqueue(session, nil)
					w.sessions.Remove(session)
					return
				}

				session.LogError(fmt.Sprintf("[%d/%d] Session healthcheck failed. Retrying in %d seconds", retryCount, maxRetries, healthcheck.RetryInterval))
			} else {
				status := session.GetStatus()
				if status == models.SessionStatusStarting {
					session.LogInfo("Session available")
				}
				if status != models.SessionStatusStarted {
					w.mediator.StartSession.Chan <- session
				}
				retryCount = 0
			}

			time.Sleep(time.Duration(healthcheck.RetryInterval) * time.Second)

		}
	}()
}
