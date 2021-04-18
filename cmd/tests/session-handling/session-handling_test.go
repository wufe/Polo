package session_handling

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/wufe/polo/internal/tests"
	"github.com/wufe/polo/pkg/models"
)

func Test_SessionHandling(t *testing.T) {

	applications := tests.Fixture(&models.RootConfiguration{
		Global: models.GlobalConfiguration{
			SessionsFolder: os.Getenv("GO_CWD") + "/.sessions",
		},
		ApplicationConfigurations: []*models.ApplicationConfiguration{
			&models.ApplicationConfiguration{
				SharedConfiguration: models.SharedConfiguration{
					Remote: "https://github.com/wufe/polo-testserver",
					Commands: models.Commands{
						Start: []models.Command{},
						Stop:  []models.Command{},
					},
				},
				Name:      "TestServer",
				IsDefault: true,
			},
		},
	})
	firstApplicationBus := applications[0].GetEventBus()

	initialized := false

	var wg sync.WaitGroup
	wg.Add(1)
	go func(ch <-chan models.ApplicationEvent) {
		defer wg.Done()
		for {
			select {
			case ev, ok := <-ch:
				if !ok {
					return
				}
				switch ev.EventType {
				case models.ApplicationEventTypeFetchCompleted:
					initialized = true
					return
				}
			case <-time.After(10 * time.Second):
				initialized = false
				return
			}
		}
	}(firstApplicationBus.GetChan())
	wg.Wait()

	if !initialized {
		t.Errorf("expected application to get initialized, but timeout was reached")
	}
}
