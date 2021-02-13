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

		retryCount := 0
		maxRetries := session.Application.Healthcheck.MaxRetries

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
			target.Path = path.Join(target.Path, session.Application.Healthcheck.URL)
			client := &http.Client{
				Timeout: time.Duration(session.Application.Healthcheck.RetryTimeout) * time.Second,
			}
			req, err := http.NewRequest(
				session.Application.Healthcheck.Method,
				target.String(),
				nil,
			)
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(session.Application.Healthcheck.RetryTimeout)*time.Second)
			req.WithContext(ctx)
			err = session.Application.Headers.ApplyTo(req)
			if err != nil {
				log.Errorf("Error applying headers to the request: %s", err.Error())
			}
			if session.Application.Host != "" {
				req.Header.Add("Host", session.Application.Host)
				req.Host = session.Application.Host
			}
			if err != nil {
				log.Errorln("Could not build HTTP request", req)
			}
			response, err := client.Do(req)
			cancel()
			if err != nil || response.StatusCode != session.Application.Healthcheck.Status {
				retryCount++

				if session.Status == models.SessionStatusStarted {
					log.Errorf("\t[S:%s] Session health degraded", session.UUID)
					session.SetStatus(models.SessionStatusDegraded)
				}
				if retryCount >= maxRetries {

					if session.GetStatus() == models.SessionStatusStarting {
						session.SetKillReason(models.KillReasonHealthcheckFailed)
					}

					log.Errorf("\t[S:%s] Session healthcheck failed. Destroying session", session.UUID)
					w.mediator.DestroySession.Enqueue(session, nil)
					w.sessions.Remove(session)
					return
				}

				log.Errorf("\t[S:%s][%d/%d] Session healthcheck failed. Retrying in %d seconds", session.UUID, retryCount, maxRetries, session.Application.Healthcheck.RetryInterval)
			} else {
				if session.Status == models.SessionStatusStarting {
					log.Infof("\t[S:%s] Session available", session.UUID)
				}
				if session.Status != models.SessionStatusStarted {
					w.mediator.StartSession.Chan <- session
				}
				retryCount = 0
			}

			time.Sleep(time.Duration(session.Application.Healthcheck.RetryInterval) * time.Second)

		}
	}()
}
