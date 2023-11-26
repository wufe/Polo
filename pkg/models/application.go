package models

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/uuid"
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models/output"
	"github.com/wufe/polo/pkg/utils"
)

var (
	ApplicationStatusLoading ApplicationStatus = "loading"
	ApplicationStatusReady   ApplicationStatus = "ready"
)

type Application struct {
	utils.RWLocker          `json:"-"`
	Filename                string `json:"filename"`
	configuration           ApplicationConfiguration
	Status                  ApplicationStatus         `json:"status"`
	Folder                  string                    `json:"folder"`
	BaseFolder              string                    `json:"baseFolder"`
	ObjectsToHashMap        map[string]string         `json:"-"`
	HashToObjectsMap        map[string]*RemoteObject  `json:"-"`
	BranchesMap             map[string]*Branch        `json:"branchesMap"`
	TagsMap                 map[string]*Tag           `json:"tagsMap"`
	Commits                 []string                  `json:"-"`
	CommitMap               map[string]*object.Commit `json:"-"`
	CompiledForwardPatterns []CompiledForwardPattern  `json:"-"`
	L4Forwards              []applicationL4Forward    `json:"-"`
	notifications           []ApplicationNotification
	bus                     *ApplicationEventBus
	log                     logging.Logger
}

type ApplicationStatus string

type CompiledForwardPattern struct {
	Pattern *regexp.Regexp
	Forward Forward
}

type applicationL4Forward struct {
	protocol        string
	sourceHost      string
	sourcePort      string
	destinationHost string
	destinationPort string
}

type ApplicationCommand struct {
	exec.Cmd
	// TODO: Add "suppressOutputPrint"
}

type ApplicationCommandOutput struct {
	Output   []string
	ExitCode int
}

type RemoteObject struct {
	Branches []string
	Tags     []string
}

type CheckoutObject struct {
	Name        string `json:"name"`
	Hash        string `json:"hash"`
	Author      string `json:"author"`
	AuthorEmail string
	Date        time.Time `json:"date"`
	Message     string    `json:"message"`
}

type Tag struct {
	CheckoutObject
}

type Branch struct {
	CheckoutObject
}

func newApplication(
	configuration *ApplicationConfiguration,
	filename string,
	mutexBuilder utils.MutexBuilder,
	logger logging.Logger,
) (*Application, error) {
	application := &Application{
		Filename: filename,
		RWLocker: mutexBuilder(),
		Status:   ApplicationStatusLoading,
		bus:      NewApplicationEventBus(mutexBuilder, logger),
		log:      logger,
	}
	configuration, err := NewApplicationConfiguration(configuration, mutexBuilder)
	if err != nil {
		return nil, err
	}
	compiled, err := compileForwardsWithPattern(configuration.Forwards)
	if err != nil {
		return nil, err
	}
	l4Forwards, err := initL4Forwards(configuration.Forwards)
	if err != nil {
		return nil, err
	}
	application.CompiledForwardPatterns = compiled
	application.L4Forwards = l4Forwards
	application.ObjectsToHashMap = make(map[string]string)
	application.HashToObjectsMap = make(map[string]*RemoteObject)
	application.BranchesMap = make(map[string]*Branch)
	application.TagsMap = make(map[string]*Tag)
	application.Commits = []string{}
	application.CommitMap = make(map[string]*object.Commit)
	if application.notifications == nil {
		application.notifications = []ApplicationNotification{}
	}
	application.SetConfiguration(*configuration)
	return application, nil
}

func compileForwardsWithPattern(forwards []Forward) ([]CompiledForwardPattern, error) {
	compiled := []CompiledForwardPattern{}
	for i, forward := range forwards {
		switch strings.ToLower(forward.Protocol) {
		case "http", "https":
			compiledPattern, err := regexp.Compile(forward.Pattern)
			if err != nil {
				return nil, fmt.Errorf("application.forwards[%d].pattern is not a valid regex: %s", i, err.Error())
			}
			compiled = append(
				compiled,
				CompiledForwardPattern{
					compiledPattern,
					forward,
				},
			)
		}
	}
	return compiled, nil
}

func initL4Forwards(forwards []Forward) ([]applicationL4Forward, error) {
	initialized := []applicationL4Forward{}
	for _, forward := range forwards {
		switch strings.ToLower(forward.Protocol) {
		case "tcp", "udp":
			initialized = append(initialized, applicationL4Forward{
				protocol:        strings.ToLower(forward.Protocol),
				sourceHost:      forward.SourceHost,
				sourcePort:      forward.SourcePort,
				destinationHost: forward.DestinationHost,
				destinationPort: forward.DestinationPort,
			})
		}
	}
	return initialized, nil
}

// ToOutput converts this model into an output model
func (a *Application) ToOutput() output.Application {
	return *mapApplication(a)
}

func (a *Application) WithLock(f func(*Application)) {
	a.Lock()
	defer a.Unlock()
	f(a)
}

func (a *Application) WithRLock(f func(*Application)) {
	a.RLock()
	defer a.RUnlock()
	f(a)
}

func (a *Application) SetFolder(folder string) {
	a.Lock()
	defer a.Unlock()
	a.Folder = folder
}

func (a *Application) SetBaseFolder(baseFolder string) {
	a.Lock()
	defer a.Unlock()
	a.BaseFolder = baseFolder
}

func (a *Application) GetStatus() ApplicationStatus {
	a.RLock()
	defer a.RUnlock()
	return a.Status
}

func (a *Application) SetStatus(status ApplicationStatus) {
	a.Lock()
	defer a.Unlock()
	a.Status = status
}

func (a *Application) GetConfiguration() ApplicationConfiguration {
	a.log.Trace("Getting application configuration")
	a.RLock()
	defer a.RUnlock()
	return a.configuration
}

func (a *Application) SetConfiguration(conf ApplicationConfiguration) {
	a.Lock()
	defer a.Unlock()
	a.configuration = conf
}

func (a *Application) GetEventBus() *ApplicationEventBus {
	a.log.Trace("Getting application event bus")
	a.RLock()
	defer a.RUnlock()
	return a.bus
}

// AddNotification adds a notification to the array of notifications
// to be shown on the client side
// The notifications are grouped by type, so only the first one gets notified
// to the final user.
func (a *Application) AddNotification(notificationType ApplicationNotificationType, description string, level ApplicationNotificationLevel, permanent bool) {
	a.Lock()
	defer a.Unlock()
	a.notifications = append(a.notifications, ApplicationNotification{
		UUID:        uuid.NewString(),
		Type:        notificationType,
		Permanent:   permanent,
		Level:       level,
		Description: description,
		CreatedAt:   time.Now(),
	})
}

// RemoveNotificationByType removes all notifications by a type
func (a *Application) RemoveNotificationByType(notificationType ApplicationNotificationType) {
	a.Lock()
	defer a.Unlock()
	notifications := []ApplicationNotification{}
	for _, notification := range a.notifications {
		if notification.Type != notificationType {
			notifications = append(notifications, notification)
		}
	}
	a.notifications = notifications
}
