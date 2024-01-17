package models

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5/plumbing/object"
	integrations_models "github.com/wufe/polo/pkg/integrations/models"
	"github.com/wufe/polo/pkg/logging"
	"github.com/wufe/polo/pkg/models/output"
	"github.com/wufe/polo/pkg/utils"
)

const (
	// SessionStatusStarting - When the session is being built
	SessionStatusStarting SessionStatus = "starting"
	// SessionStatusStarted - When the session has been built
	// and the session is available to be proxied to
	SessionStatusStarted SessionStatus = "started"
	// SessionStatusStartFailed - When the session build process failed
	SessionStatusStartFailed SessionStatus = "start_failed"
	// SessionStatusStopFailed - When the session stop process failed
	SessionStatusStopFailed SessionStatus = "stop_failed"
	// SessionStatusStopping - When the session is being stopped
	SessionStatusStopping SessionStatus = "stopping"
	// SessionStatusStopped - When the session has been stopped successfully
	SessionStatusStopped SessionStatus = "stopped"
	// SessionStatusDegraded - When the healthcheck failed
	// and the session is NOT available to be proxied to
	SessionStatusDegraded SessionStatus = "degraded"

	// LogTypeStdin is the command being executed
	LogTypeStdin LogType = "stdin"
	// LogTypeStdout is the output printed on the stdout
	LogTypeStdout LogType = "stdout"
	// LogTypeStderr is the output printed on the stderr
	LogTypeStderr LogType = "stderr"

	// KillReasonNone - The reason has not been set. Maybe because there has not been a kill yet
	KillReasonNone KillReason = ""
	// KillReasonStopped - The session has been manually stopped by the user
	KillReasonStopped KillReason = "stopped"
	// KillReasonBuildFailed - The session has been killed because its build process failed
	KillReasonBuildFailed KillReason = "build_failed"
	// KillReasonHealthcheckFailed - The session has been killed because the healthcheck process
	// could not check the service reachability. It depends on user-provided configuration
	KillReasonHealthcheckFailed KillReason = "healthcheck_failed"
	// KillReasonReplaced - Means the session is going to be replaced with an updated one
	KillReasonReplaced KillReason = "replaced"

	// SessionBuildContextKey is the name of the shared BUILD context.
	// It is shared to allow an early session destruction to stop a running build of a session
	SessionBuildContextKey string = "build"
)

var (
	stdinColor    = color.New(color.BgHiCyan, color.FgBlack)
	traceColor    = color.New(color.FgHiMagenta)
	debugColor    = color.New(color.BgHiBlue, color.FgWhite)
	infoColor     = color.New(color.FgBlack)
	warnColor     = color.New(color.FgHiYellow)
	errorColor    = color.New(color.FgHiRed)
	criticalColor = color.New(color.BgHiRed, color.FgBlack)
	pointerColor  = color.New(color.FgHiBlack)
)

// SessionStatus is the status of the session
type SessionStatus string

func (status SessionStatus) String() string {
	return string(status)
}

// IsAlive states whether the session is started or about be started
func (status SessionStatus) IsAlive() bool {
	return status != SessionStatusStartFailed &&
		status != SessionStatusStopFailed &&
		status != SessionStatusStopped &&
		status != SessionStatusStopping
}

// KillReason states the reason why a session has been killed
type KillReason string

// PreventsRebuild checks if the reason why this session has been killed
// prevents another build with the same checkout to be built
func (reason KillReason) PreventsRebuild() bool {
	return reason == KillReasonStopped ||
		reason == KillReasonBuildFailed
}

// Session is a process on which an application is available.
// When a session is started it gets built starting from a branch,
// and when all is ready the reverse proxy will start pointing to it.
type Session struct {
	utils.RWLocker  `json:"-"`
	UUID            string       `json:"uuid"`
	Alias           string       `json:"alias"`
	DisplayName     string       `json:"displayName"`
	Port            int          `json:"port"`
	ApplicationName string       `json:"applicationName"`
	Application     *Application `json:"-"`
	configuration   ApplicationConfiguration
	Status          SessionStatus                `json:"status"`
	CommitID        string                       `json:"commitID"` // The object to be checked out (branch/tag/commit id)
	Checkout        string                       `json:"checkout"`
	Commit          object.Commit                `json:"commit"`
	Folder          string                       `json:"folder"`
	Variables       Variables                    `json:"variables"`
	Metrics         []Metric                     `json:"metrics"`
	Integrations    *integrations_models.Session `json:"integrations"`
	Context         *contextStore                `json:"-"`
	L4Forwards      []*L4Forward                 `json:"-"`
	logs            []Log
	shortUUID       string
	createdAt       time.Time
	inactiveAt      time.Time
	maxAge          int
	startupRetries  int
	killReason      KillReason
	// If set, states that this session replaces a previous one
	replaces               []*Session
	replacedBy             *Session
	diagnostics            []DiagnosticsData
	bus                    *SessionLifetimeEventBus
	log                    logging.Logger
	tty                    utils.TTYOutput
	advancedTerminalOutput bool
}

// L4Forward is a struct which represents a Level 4 forwarding
// A forwarding should get activated when the healthcheck has finished successfully
// and the traffic needs to be forwarded through these protocols
type L4Forward struct {
	sync.RWMutex
	IsActive        bool
	Forwarder       utils.Forwarder
	Protocol        string
	SourceHost      string
	SourcePort      string
	DestinationHost string
	DestinationPort string
}

func (f *L4Forward) Activate(onClientConnectCallback func(address string)) error {
	f.Lock()
	defer f.Unlock()

	if strings.ContainsAny(f.SourceHost, "{}") {
		return fmt.Errorf("cannot use %s as source host: placeholder not yet resolved", f.SourceHost)
	}

	if strings.ContainsAny(f.SourcePort, "{}") {
		return fmt.Errorf("cannot use %s as source port: placeholder not yet resolved", f.SourcePort)
	}

	if strings.ContainsAny(f.DestinationHost, "{}") {
		return fmt.Errorf("cannot use %s as destination host: placeholder not yet resolved", f.DestinationHost)
	}

	if strings.ContainsAny(f.DestinationPort, "{}") {
		return fmt.Errorf("cannot use %s as destination port: placeholder not yet resolved", f.DestinationPort)
	}

	if f.IsActive {
		return nil
	}

	var forwarder utils.Forwarder
	var err error
	if strings.ToLower(f.Protocol) == "tcp" {
		forwarder, err = utils.ForwardTCP(
			fmt.Sprintf("%s:%s", f.SourceHost, f.SourcePort),
			fmt.Sprintf("%s:%s", f.DestinationHost, f.DestinationPort),
		)
	} else {
		forwarder, err = utils.ForwardUDP(
			fmt.Sprintf("%s:%s", f.SourceHost, f.SourcePort),
			fmt.Sprintf("%s:%s", f.DestinationHost, f.DestinationPort),
		)
	}

	if err != nil {
		return err
	}

	f.IsActive = true
	f.Forwarder = forwarder
	forwarder.OnConnect(onClientConnectCallback)

	return nil
}

func (f *L4Forward) Deactivate() {
	f.Lock()
	defer f.Unlock()

	if !f.IsActive {
		return
	}

	f.Forwarder.Close()
	f.IsActive = false
	f.Forwarder = nil
}

func (f *L4Forward) RefreshWithVariables(vars Variables) {
	f.SourceHost = vars.ApplyTo(f.SourceHost)
	f.SourcePort = vars.ApplyTo(f.SourcePort)
	f.DestinationHost = vars.ApplyTo(f.DestinationHost)
	f.DestinationPort = vars.ApplyTo(f.DestinationPort)
}

// Variables are those variables used by a single session.
// May contain data put by the session build process
// or the output of build commands
type Variables map[string]string

// ApplyTo allows a string with placeholders to get
// those placeholders replaced by corresponding variables
func (v Variables) ApplyTo(str string) string {
	for key, value := range v {
		str = strings.ReplaceAll(str, fmt.Sprintf("{{%s}}", key), value)
	}
	return str
}

// newSession builds a session starting from a pre-built one.
// It is useful to set variable that needs to be set at initialization time
func newSession(
	session *Session,
	mutexBuilder utils.MutexBuilder,
	logger logging.Logger,
	advancedTerminalOutput bool,
) *Session {
	session.log = logger
	session.shortUUID = strings.Split(session.UUID, "-")[0]
	session.RWLocker = mutexBuilder()
	if session.bus == nil {
		session.bus = NewSessionBuildEventBus(mutexBuilder)
	}
	if session.ApplicationName == "" {
		session.ApplicationName = session.Application.GetConfiguration().Name
	}
	session.Status = SessionStatusStarting
	if session.logs == nil {
		session.logs = []Log{}
	}
	if len(session.Variables) == 0 {
		session.Variables = make(map[string]string)
	}
	if session.Metrics == nil {
		session.Metrics = []Metric{}
	}

	session.Integrations = integrations_models.NewSession(session.Integrations)

	session.createdAt = time.Now()
	session.killReason = KillReasonNone
	session.Context = NewContextStore(mutexBuilder)
	if session.Application != nil {
		session.configuration = session.getMatchingConfiguration()

		session.InitializeL4Forwards()
	}
	if session.diagnostics == nil {
		session.diagnostics = []DiagnosticsData{}
	}
	if session.replaces == nil {
		session.replaces = []*Session{}
	}

	session.tty = utils.NewTTYOutput()
	session.advancedTerminalOutput = advancedTerminalOutput

	return session
}

// InitializeL4Forwards gets called when a session gets built or
// when a session has already been built, like when Polo shuts down,
// a session is still alive, and when it goes up again, the session gets resumed.
// In this case the newSession gets called but no application has already been associated
// so we try to initialize the level 4 forwards a second time.
func (session *Session) InitializeL4Forwards() {
	if len(session.L4Forwards) == 0 {
		for _, f := range session.Application.L4Forwards {
			session.L4Forwards = append(session.L4Forwards, &L4Forward{
				IsActive:        false,
				Protocol:        f.protocol,
				SourceHost:      f.sourceHost,
				SourcePort:      f.sourcePort,
				DestinationHost: f.destinationHost,
				DestinationPort: f.destinationPort,
			})
		}
	}
}

func (session *Session) GetCreatedAt() time.Time {
	session.RLock()
	defer session.RUnlock()
	return session.createdAt
}

// ToOutput converts this model into an output model
func (session *Session) ToOutput() output.Session {
	return *MapSession(session)
}

// SetReplaces thread-safely sets the session which will be replaced
// when this session will go online
func (session *Session) SetReplaces(replaces []*Session) {
	session.log.Trace("Setting replaces")
	session.Lock()
	defer session.Unlock()
	if replaces == nil {
		session.replaces = []*Session{}
	} else {
		session.replaces = replaces
	}
}

// GetReplaces thread-safely retrieves the session which will be replaced
// when this session will go online
func (session *Session) GetReplaces() []*Session {
	session.RLock()
	defer session.RUnlock()
	return session.replaces
}

// SetReplacedBy thread-safely sets the UUID of the session by which
// this session has been replaced
func (session *Session) SetReplacedBy(newSession *Session) {
	session.Lock()
	defer session.Unlock()
	session.replacedBy = newSession
}

// GetReplacedBy thread-safely retrieves the UUID of the session by which
// this session has been replaced
func (session *Session) GetReplacedBy() *Session {
	session.RLock()
	defer session.RUnlock()
	return session.replacedBy
}

// GetConfiguration allows to retrieve the CURRENT configuration in a thread-safe manner.
// This configuration gets replaced whenever there's an update by the user.
// So it is advisable to not store indefinitely this configuration, but to ask for it when needed
func (session *Session) GetConfiguration() ApplicationConfiguration {
	session.log.Trace("Getting session configuration")
	session.RLock()
	defer session.RUnlock()
	return session.configuration
}

// InitializeConfiguration gets called whenever a secondary actor knows
// that application's configuration changed.
// This allows the session to get its matching configuration
func (session *Session) InitializeConfiguration() {
	session.Lock()
	defer session.Unlock()
	session.configuration = session.getMatchingConfiguration()
}

// getMatchingConfiguration cycles through all available configuration overrides
// and returns the default one, with overrides applied
func (session *Session) getMatchingConfiguration() ApplicationConfiguration {
	branches := session.Application.configuration.Branches
	baseConfig := session.Application.GetConfiguration()
	if branches == nil {
		return baseConfig
	}
	checkout := session.Checkout
	found := false
	var matchingConf SharedConfiguration
	for _, conf := range branches {
		testRE := regexp.MustCompile(conf.Test)
		if testRE.MatchString(checkout) {
			matchingConf = conf.SharedConfiguration
			found = true
			break
		}
	}
	if !found {
		return baseConfig
	}
	baseConfig.OverrideWith(matchingConf)
	return baseConfig
}

// LogCritical logs a message to stdout and stores it in the session logs slice
func (session *Session) LogCritical(message []byte) {
	session.Lock()
	defer session.Unlock()

	session.log.Errorf(fmt.Sprintf("\t[%s]: %s", session.shortUUID, string(message)))

	if session.advancedTerminalOutput {
		session.logFormattedOutput(criticalColor.Sprint(string(message)))
	} else {
		session.logs = append(
			session.logs,
			NewLog(string(message), LogTypeCritical),
		)
	}
}

// LogError logs a message to stdout and stores it in the session logs slice
func (session *Session) LogError(message []byte) {
	session.Lock()
	defer session.Unlock()

	session.log.Errorf(fmt.Sprintf("\t[%s]: %s", session.shortUUID, string(message)))

	if session.advancedTerminalOutput {
		session.logFormattedOutput(errorColor.Sprint(string(message)))
	} else {
		session.logs = append(
			session.logs,
			NewLog(string(message), LogTypeError),
		)
	}
}

// LogWarn logs a message to stdout and stores it in the session logs slice
func (session *Session) LogWarn(message []byte) {
	session.Lock()
	defer session.Unlock()

	session.log.Warnf(fmt.Sprintf("\t[%s]: %s", session.shortUUID, string(message)))

	if session.advancedTerminalOutput {
		session.logFormattedOutput(warnColor.Sprint(string(message)))
	} else {
		session.logs = append(
			session.logs,
			NewLog(string(message), LogTypeWarn),
		)
	}
}

// LogInfo logs a message to stdout and stores it in the session logs slice
func (session *Session) LogInfo(message []byte) {
	session.Lock()
	defer session.Unlock()

	session.log.Infof(fmt.Sprintf("\t[%s]: %s", session.shortUUID, string(message)))

	if session.advancedTerminalOutput {
		session.logFormattedOutput(infoColor.Sprint(string(message)))
	} else {
		session.logs = append(
			session.logs,
			NewLog(string(message), LogTypeInfo),
		)
	}
}

// LogDebug logs a message to stdout and stores it in the session logs slice
func (session *Session) LogDebug(message []byte) {
	session.Lock()
	defer session.Unlock()

	session.log.Debugf(fmt.Sprintf("\t[%s]: %s", session.shortUUID, string(message)))

	if session.advancedTerminalOutput {
		session.logFormattedOutput(debugColor.Sprint(string(message)))
	} else {
		session.logs = append(
			session.logs,
			NewLog(string(message), LogTypeDebug),
		)
	}
}

// LogTrace logs a message to stdout and stores it in the session logs slice
func (session *Session) LogTrace(message []byte) {
	session.Lock()
	defer session.Unlock()

	session.log.Tracef(fmt.Sprintf("\t[%s]: %s", session.shortUUID, string(message)))

	if session.advancedTerminalOutput {
		session.logFormattedOutput(traceColor.Sprint(string(message)))
	} else {
		session.logs = append(
			session.logs,
			NewLog(string(message), LogTypeTrace),
		)
	}
}

// LogStdin logs a message to stdout and stores it in the session logs slice
func (session *Session) LogStdin(message []byte) {
	session.Lock()
	defer session.Unlock()
	session.log.Infof(fmt.Sprintf("\t\t[%s (stdin)>]: %s", session.shortUUID, string(message)))

	if session.advancedTerminalOutput {
		session.logFormattedOutput(stdinColor.Sprint(string(message)))
	} else {
		session.logs = append(
			session.logs,
			NewLog(string(message), LogTypeStdin),
		)
	}
}

func (session *Session) logFormattedOutput(message string) error {

	formatted := append(
		[]byte(pointerColor.Sprintf("[%s] ▶ ", time.Now().Format(time.TimeOnly))),
		append(
			[]byte(message),
			[]byte{'\r', '\n'}...,
		)...,
	)

	_, err := session.tty.Write(formatted)
	return err
}

// LogStdout logs a message to stdout and stores it in the session logs slice
func (session *Session) LogStdout(message []byte) {
	session.Lock()
	defer session.Unlock()
	session.log.Infof(fmt.Sprintf("\t\t[%s (stdout)>]: %s", session.shortUUID, string(message)))

	if session.advancedTerminalOutput {
		session.tty.Write(append(message, '\r', '\n'))
	} else {
		session.logs = append(
			session.logs,
			NewLog(string(message), LogTypeStdout),
		)
	}
}

// LogStderr logs a message to stdout and stores it in the session logs slice
func (session *Session) LogStderr(message []byte) {
	session.Lock()
	defer session.Unlock()
	session.log.Infof(fmt.Sprintf("\t\t[%s (stderr)>]: %s", session.shortUUID, string(message)))

	if session.advancedTerminalOutput {
		session.tty.Write(append(message, '\r', '\n'))
	} else {
		session.logs = append(
			session.logs,
			NewLog(string(message), LogTypeStderr),
		)
	}
}

// MarkAsBeingRequested informs the session that it has been used by a proxy
// so it must reset its inactivity timer, if available
func (session *Session) MarkAsBeingRequested() {
	conf := session.GetConfiguration()
	if session.GetMaxAge() > -1 {
		// Refreshes the inactiveAt field every time someone makes a request to this session
		session.SetInactiveAt(time.Now().Add(time.Second * time.Duration(conf.Recycle.InactivityTimeout)))
		session.SetMaxAge(conf.Recycle.InactivityTimeout)
	}
}

// SetStatus allows to set the session status thread-safely
func (session *Session) SetStatus(status SessionStatus) {
	session.Lock()
	session.log.Tracef("[%s]: Updated status from %s to %s", session.shortUUID, session.Status, status)
	defer session.Unlock()
	previousStatus := session.Status
	session.Status = status
	session.diagnostics = append(session.diagnostics, DiagnosticsData{
		Action: DiagnosticsActionReplacement,
		When:   time.Now(),
		Field:  "status",
		Value: PrevNextDiagnosticsValue{
			Previous: previousStatus.String(),
			Next:     status.String(),
		},
	})
}

// GetStatus allows to get the session status thread-safely
func (session *Session) GetStatus() SessionStatus {
	session.RLock()
	defer session.RUnlock()
	return session.Status
}

// DecreaseMaxAge decreases the max-age parameter of the session thread-safely
func (session *Session) DecreaseMaxAge() {
	session.Lock()
	defer session.Unlock()
	session.maxAge--
}

// GetMaxAge allows to retrieve the session max-age thread-safely
func (session *Session) GetMaxAge() int {
	session.RLock()
	defer session.RUnlock()
	return session.maxAge
}

// SetMaxAge allows to set an exact max-age value for the session thread-safely
func (session *Session) SetMaxAge(age int) {
	session.Lock()
	defer session.Unlock()
	session.maxAge = age
}

// GetInactiveAt retrieves the inactive-at value for the session thread-safely
func (session *Session) GetInactiveAt() time.Time {
	session.RLock()
	defer session.RUnlock()
	return session.inactiveAt
}

// SetInactiveAt is the thread-safe setter for InactiveAt
func (session *Session) SetInactiveAt(at time.Time) {
	session.Lock()
	defer session.Unlock()
	session.inactiveAt = at
}

// GetStartupRetriesCount retrieves the current count of startup retries thread-safely
func (session *Session) GetStartupRetriesCount() int {
	session.RLock()
	defer session.RUnlock()
	return session.startupRetries
}

// IncStartupRetriesCount thread-safely increments the current count of startup retries
func (session *Session) IncStartupRetriesCount() {
	session.log.Trace("Incrementing startup retries count")
	session.Lock()
	defer session.Unlock()
	session.startupRetries++
}

// ResetStartupRetriesCount thread-safely resets the current count of startup retries
func (session *Session) ResetStartupRetriesCount() {
	session.Lock()
	defer session.Unlock()
	session.startupRetries = 0
}

// GetKillReason returns the reason why the session has been killed thread-safely.
// Returns KillReasonNone if the session has not been killed
func (session *Session) GetKillReason() KillReason {
	session.log.Trace("Getting kill reason")
	session.RLock()
	defer session.RUnlock()
	return session.killReason
}

// SetKillReason allows to set the session kill reason thread-safely
func (session *Session) SetKillReason(reason KillReason) {
	session.Lock()
	defer session.Unlock()
	session.killReason = reason
}

// SetVariable thread-safely sets a variable value into the session variables dictionary
func (session *Session) SetVariable(k string, v string) {
	session.Lock()
	defer session.Unlock()
	session.Variables[k] = v
}

// ResetVariables thread-safely resets the session variables dictionary
func (session *Session) ResetVariables() {
	session.log.Trace("Resetting variables")
	session.Lock()
	defer session.Unlock()
	session.Variables = make(map[string]string)
}

// IsAlive thread-safely retrieves if the session is alive or not
func (session *Session) IsAlive() bool {
	session.RLock()
	defer session.RUnlock()
	return session.Status.IsAlive()
}

func (session *Session) GetLogs() []Log {
	session.RLock()
	defer session.RUnlock()
	return session.logs
}

func (session *Session) GetDiagnosticsData() []DiagnosticsData {
	session.RLock()
	defer session.RUnlock()
	return session.diagnostics
}

func (session *Session) GetEventBus() *SessionLifetimeEventBus {
	session.log.Trace("Getting session event bus")
	session.RLock()
	defer session.RUnlock()
	return session.bus
}

func (session *Session) GetTarget() string {
	session.RLock()
	defer session.RUnlock()
	return session.getTargetInternal()
}

func (session *Session) getTargetInternal() string {
	target := session.configuration.Target
	return session.Variables.ApplyTo(target)
}

func (session *Session) GetTTYOutput() utils.TTYOutput {
	session.RLock()
	defer session.RUnlock()
	return session.tty
}
