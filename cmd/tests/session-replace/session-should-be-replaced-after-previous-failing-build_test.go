package session_replace

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/logrusorgru/aurora/v3"
	"github.com/wufe/polo/internal/tests"
	"github.com/wufe/polo/internal/tests/events_assertions"
	"github.com/wufe/polo/internal/tests/execution_fixture"
	"github.com/wufe/polo/internal/tests/net_fixture"
	"github.com/wufe/polo/internal/tests/versioning_fixture"
	"github.com/wufe/polo/pkg/models"
)

// Session should be replaced, even if the previous session builds failed
func Test_SessionShouldBeReplacedAfterPreviousFailingBuild(t *testing.T) {

	// Create the HTTP server and start it
	httpServer := net_fixture.NewHTTPServerFixture()
	port, tearDown := httpServer.Setup()
	defer tearDown()

	// Create port retriever to set up HTTP server port
	portRetriever := net_fixture.NewPortRetrieverFixture()
	portRetriever.SetFreePort(port)

	// Create command runner fixture
	commandRunner := execution_fixture.NewCommandRunnerFixture()

	fetcher := versioning_fixture.NewRepositoryFetcher()

	// Creating the branch
	branch := fetcher.NewBranch("main")

	// Creating the first commit
	firstCommit := fetcher.NewCommit("First commit")
	fetcher.AddCommitToBranch(firstCommit, branch)

	// Setup the application
	di := tests.Fixture(&tests.InjectableServices{
		RepositoryFetcher: fetcher,
		GitClient:         versioning_fixture.NewGitClient(),
		CommandRunner:     commandRunner,
		PortRetriever:     portRetriever,
	}, models.BuildApplicationConfiguration("Test_SessionShouldBeReplacedAfterPreviousFailingBuild").
		WithRemote("FakeRemote").
		WithStartCommand("valid-command.exe").
		WithStopCommand("valid-command.exe").
		WithStartupRetries(3).
		WithHealthcheckRetryInterval(1).
		SetAsDefault(true).
		WithBranch(
			models.BuildBranchConfigurationMatch("main").
				SetWatch(false).
				SetMain(false),
		),
	)

	// Get events channel
	applications := di.GetApplications()
	firstApplication := applications[0]
	firstApplicationBus := firstApplication.GetEventBus()
	firstApplicationChan := firstApplicationBus.GetChan()

	// Assert application is being loaded
	events_assertions.AssertApplicationGetsInitializedAndFetched(firstApplicationChan, t)

	// Creating the second commit
	secondCommit := fetcher.NewCommit("Second commit")
	fetcher.AddCommitToBranch(secondCommit, branch)

	// Re-fetch the application
	mediator := di.GetMediator()
	mediator.ApplicationFetch.Enqueue(firstApplication, false)

	// Assert application gets fetched
	events_assertions.AssertApplicationGetsFetched(firstApplicationChan, t)

	// Request new session to be built
	requestService := di.GetRequestService()
	sessionBuildResult, err := requestService.NewSession(branch.Name, firstApplication.GetConfiguration().Name)
	if err != nil {
		t.Error(err.Error())
	}

	// Assert application session gets built
	events_assertions.AssertApplicationSessionSucceeded(firstApplicationChan, t)

	// Get session events channel
	session := sessionBuildResult.Session
	sessionBus := session.GetEventBus()
	sessionChan := sessionBus.GetChan()

	// Assert session fails to get created
	events_assertions.AssertSessionEvents(
		sessionChan,
		[]models.SessionEventType{
			// First build
			models.SessionEventTypeBuildStarted,
			models.SessionEventTypePreparingFolders,
			models.SessionEventTypeCommandsExecutionStarted,
			models.SessionEventTypeHealthcheckStarted,
			models.SessionEventTypeHealthcheckSucceded,
			models.SessionEventTypeSessionAvailable,
			models.SessionEventTypeSessionStarted,
		},
		t,
		10*time.Second,
	)

	// Tell the command runner fixture to fail next command, and next 3 retries
	commandRunner.FailNextNCommands(4)

	// Creating the third commit
	thirdCommit := fetcher.NewCommit("Third commit")
	fetcher.AddCommitToBranch(thirdCommit, branch)

	// Re-fetch the application
	mediator.ApplicationFetch.Enqueue(firstApplication, true)

	// Assert application gets fetched and build fails
	events_assertions.AssertApplicationGetsFetchedWithHotSwapAndFailingBuildWith3Retries(firstApplicationChan, t)

	sessionStorage := di.GetSessionStorage()
	aliveSessions := sessionStorage.GetAllAliveSessions()

	if len(aliveSessions) > 1 {

		t.Errorf(aurora.Sprintf(aurora.Red("expected number of alive sessions to be 1, but found %d"), len(aliveSessions)))
		for i, s := range aliveSessions {
			statusOutput, _ := json.MarshalIndent(s.ToOutput(), "", "    ")

			t.Errorf(aurora.Sprintf(aurora.Red("session #%d:\n%s"), i, aurora.Cyan((statusOutput))))
		}
		return
	}

	// Creating the fourth commit
	fourthCommit := fetcher.NewCommit("Fourth commit")
	fetcher.AddCommitToBranch(fourthCommit, branch)

	// Re-fetch the application
	mediator.ApplicationFetch.Enqueue(firstApplication, true)

	// Assert application gets fetched and build succeeds
	events_assertions.AssertApplicationEvents(
		firstApplicationChan,
		[]models.ApplicationEventType{
			models.ApplicationEventTypeFetchStarted,
			models.ApplicationEventTypeHotSwap,
			models.ApplicationEventTypeSessionBuild,
			models.ApplicationEventTypeFetchCompleted,
			models.ApplicationEventTypeSessionBuildSucceeded,
		},
		t,
		2*time.Second,
	)

	// Assert alive session is just one
	aliveSessions = sessionStorage.GetAllAliveSessions()

	lastAliveSession := aliveSessions[len(aliveSessions)-1]
	lastAliveSessionChan := lastAliveSession.GetEventBus().GetChan()

	events_assertions.AssertSessionEvents(
		lastAliveSessionChan,
		[]models.SessionEventType{
			models.SessionEventTypeBuildStarted,
			models.SessionEventTypePreparingFolders,
			models.SessionEventTypeCommandsExecutionStarted,
			models.SessionEventTypeHealthcheckStarted,
			models.SessionEventTypeHealthcheckSucceded,
			models.SessionEventTypeSessionAvailable,
			models.SessionEventTypeSessionStarted,
		},
		t,
		2*time.Second,
	)

	aliveSessions = sessionStorage.GetAllAliveSessions()

	if len(aliveSessions) > 1 {

		t.Errorf(aurora.Sprintf(aurora.Red("expected number of alive sessions to be 1, but found %d"), len(aliveSessions)))
		for i, s := range aliveSessions {
			statusOutput, _ := json.MarshalIndent(s.ToOutput(), "", "    ")

			t.Errorf(aurora.Sprintf(aurora.Red("session #%d:\n%s"), i, aurora.Cyan((statusOutput))))
		}
		return
	}

}
