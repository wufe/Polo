package background

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/models"
)

type SessionHealthcheckWorker struct {
	sessions *threadSafeSlice
	mediator *Mediator
}

type threadSafeSlice struct {
	sync.Mutex
	elements []*models.Session
}

func (slice *threadSafeSlice) Push(s *models.Session) {
	slice.Lock()
	defer slice.Unlock()

	slice.elements = append(slice.elements, s)
}

func (slice *threadSafeSlice) Find(s *models.Session) *models.Session {
	slice.Lock()
	defer slice.Unlock()

	var foundSession *models.Session
	for _, session := range slice.elements {
		if s == session {
			foundSession = s
			break
		}
	}

	return foundSession
}

func (slice *threadSafeSlice) Remove(s *models.Session) {
	slice.Lock()
	defer slice.Unlock()

	index := -1
	for i, session := range slice.elements {
		if session == s {
			index = i
		}
	}
	if index > -1 {
		slice.elements = append(slice.elements[:index], slice.elements[index+1:]...)
	}
}

func NewSessionHealthcheckWorker(
	mediator *Mediator,
) *SessionHealthcheckWorker {
	worker := &SessionHealthcheckWorker{
		sessions: &threadSafeSlice{
			elements: []*models.Session{},
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
			}
			w.mediator.HealthcheckSession.ResponseChan <- struct{}{}
		}
	}()
}

func (w *SessionHealthcheckWorker) startHealthchecking(session *models.Session) {
	go func() {
		time.Sleep(5 * time.Second)

		// TODO: Use this
		retryCount := 0
		maxRetries := session.Application.Healthcheck.MaxRetries

		for {
			// Failed or destroyed
			if !session.Status.IsAlive() {
				w.sessions.Remove(session)
				return
			}

			target, err := url.Parse(session.Target)
			if err != nil {
				session.LogError(fmt.Sprintf("Could not parse target URL: %s", err.Error()))
				log.Errorln("Could not parse target URL", err)
				w.mediator.DestroySession.Request(session, nil)
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
					log.Errorf("[SESSION:%s] Session health degraded", session.UUID)
					session.Status = models.SessionStatusDegraded
				}
				if retryCount >= maxRetries {
					log.Errorf("[SESSION:%s] Session healthcheck failed. Destroying session", session.UUID)
					w.mediator.DestroySession.Request(session, nil)
					w.sessions.Remove(session)
					return
				}

				log.Errorf("[SESSION:%s][%d/%d] Session healthcheck failed. Retrying in %d seconds", session.UUID, retryCount, maxRetries, session.Application.Healthcheck.RetryInterval)
			} else {
				if session.Status == models.SessionStatusStarting {
					log.Infof("[SESSION:%s] Session available", session.UUID)
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