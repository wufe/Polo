package background

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/wufe/polo/pkg/background/queues"
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/utils"
)

type SessionHealthcheckWorker struct {
	sessions *utils.ThreadSafeSlice
	mediator *Mediator
	log      logging.Logger
}

func NewSessionHealthcheckWorker(
	mediator *Mediator,
	logger logging.Logger,
) *SessionHealthcheckWorker {
	worker := &SessionHealthcheckWorker{
		sessions: &utils.ThreadSafeSlice{
			Elements: []interface{}{},
		},
		mediator: mediator,
		log:      logger,
	}
	return worker
}

func (w *SessionHealthcheckWorker) Start() {
	w.startAcceptingSessionHealthcheckingRequests()
}

func (w *SessionHealthcheckWorker) startAcceptingSessionHealthcheckingRequests() {
	go func() {
		for {
			request := <-w.mediator.HealthcheckSession.RequestChan
			foundSession := w.sessions.Find(request.Session)
			if foundSession == nil {
				w.startHealthchecking(request.Session)
			} else {
				w.log.Errorln("ALREADY THERE")
			}
			w.mediator.HealthcheckSession.ResponseChan <- struct{}{}
		}
	}()
}

func (w *SessionHealthcheckWorker) startHealthchecking(session *models.Session) {
	session.GetEventBus().PublishEvent(models.SessionEventTypeHealthcheckStarted, session)
	w.sessions.Push(session)
	go func() {

		conf := session.GetConfiguration()
		maxRetries := conf.Healthcheck.MaxRetries
		healthcheck := conf.Healthcheck
		headers := conf.Headers
		host := conf.Host

		retryCount := 0

		time.Sleep(time.Duration(healthcheck.RetryInterval) * time.Second)

		for {
			// Failed or destroyed
			if !session.GetStatus().IsAlive() {
				w.sessions.Remove(session)
				return
			}

			target, err := url.Parse(session.GetTarget())
			if err != nil {
				session.LogError([]byte(fmt.Sprintf("Could not parse target URL: %s", err.Error())))
				w.log.Errorln("Could not parse target URL", err)
				w.mediator.DestroySession.Enqueue(session, nil)
				w.sessions.Remove(session)
				return
			}
			target.Path = path.Join(target.Path, healthcheck.URL)
			client := &http.Client{
				Timeout: time.Duration(healthcheck.Timeout) * time.Second,
			}
			req, err := http.NewRequest(
				healthcheck.Method,
				target.String(),
				nil,
			)
			if err != nil {
				w.log.Errorln("Could not build HTTP request", req)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(healthcheck.Timeout)*time.Second)
			req.WithContext(ctx)
			err = headers.ApplyTo(req)
			if err != nil {
				w.log.Errorf("Error applying headers to the request: %s", err.Error())
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
					session.LogWarn([]byte("Session health degraded"))
					session.SetStatus(models.SessionStatusDegraded)
				}
				if retryCount >= maxRetries {

					if session.GetStatus() == models.SessionStatusStarting {
						session.SetKillReason(models.KillReasonHealthcheckFailed)
					}

					session.LogError([]byte("Session healthcheck failed. Destroying session"))
					w.mediator.DestroySession.Enqueue(session, nil)
					w.sessions.Remove(session)
					session.GetEventBus().PublishEvent(models.SessionEventTypeHealthcheckFailed, session)
					session.GetEventBus().Close()
					return
				}

				session.LogError([]byte(fmt.Sprintf("[%d/%d] Session healthcheck failed. Retrying in %.2f seconds", retryCount, maxRetries, healthcheck.RetryInterval)))
			} else {
				status := session.GetStatus()
				if status == models.SessionStatusStarting {
					session.LogInfo([]byte("Session available"))
					session.GetEventBus().PublishEvent(models.SessionEventTypeHealthcheckSucceded, session)
					session.GetEventBus().PublishEvent(models.SessionEventTypeSessionAvailable, session)
				}
				if status != models.SessionStatusStarted {
					w.mediator.StartSession.Enqueue(queues.SessionStartInput{
						Session: session,
					})
				}
				retryCount = 0
			}

			time.Sleep(time.Duration(healthcheck.RetryInterval) * time.Second)

		}
	}()
}
