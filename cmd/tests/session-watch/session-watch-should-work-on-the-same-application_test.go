package session_watch

import (
	"testing"

	"github.com/wufe/polo/internal/tests"
	"github.com/wufe/polo/internal/tests/events_assertions"
	"github.com/wufe/polo/internal/tests/execution_fixture"
	"github.com/wufe/polo/internal/tests/net_fixture"
	"github.com/wufe/polo/internal/tests/versioning_fixture"
	"github.com/wufe/polo/pkg/models"
)

// Session watch and auto-start should be started from the same application
// which has the session already started
func Test_SessionWatchShouldWorkOnTheSameApplication(t *testing.T) {

	// Create the HTTP server and start it
	httpServer := net_fixture.NewHTTPServerFixture()
	port, tearDown := httpServer.Setup()
	defer tearDown()

	// Create port retriever to set up HTTP server port
	portRetriever := net_fixture.NewPortRetrieverFixture()
	portRetriever.SetFreePort(port)

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
		CommandRunner:     execution_fixture.NewCommandRunnerFixture(),
		PortRetriever:     portRetriever,
	}, models.BuildApplicationConfiguration("Test_SessionWatchShouldWorkOnTheSameApplication1").
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
		models.BuildApplicationConfiguration("Test_SessionWatchShouldWorkOnTheSameApplication2").
			WithRemote("FakeRemote").
			WithStartCommand("valid-command.exe").
			WithStopCommand("valid-command.exe").
			WithStartupRetries(3).
			WithHealthcheckRetryInterval(1).
			SetAsDefault(false).
			WithBranch(
				models.BuildBranchConfigurationMatch("main").
					SetWatch(false).
					SetMain(false),
			),
	)

	// Get events channel
	applications := di.GetApplications()
	firstApplication := applications[0]
	firstApplicationChan := firstApplication.GetEventBus().GetChan()
	secondApplication := applications[1]
	secondApplicationChan := secondApplication.GetEventBus().GetChan()

	// Assert first application is being loaded
	events_assertions.AssertApplicationGetsInitializedAndFetched(firstApplicationChan, t)

	// Assert second application is being loaded
	events_assertions.AssertApplicationGetsInitializedAndFetched(secondApplicationChan, t)

	// Creating the second commit
	secondCommit := fetcher.NewCommit("Second commit")
	fetcher.AddCommitToBranch(secondCommit, branch)

	// Re-fetch the first application
	mediator := di.GetMediator()
	mediator.ApplicationFetch.Enqueue(firstApplication, false)

	// Assert first application gets fetched
	events_assertions.AssertApplicationGetsFetched(firstApplicationChan, t)

	// Re-fetch the second application
	mediator.ApplicationFetch.Enqueue(secondApplication, false)

	// Assert second application gets fetched
	events_assertions.AssertApplicationGetsFetched(secondApplicationChan, t)

	// Request new session to be built
	requestService := di.GetRequestService()
	sessionBuildResult, err := requestService.NewSession(branch.Name, firstApplication.GetConfiguration().Name)
	if err != nil {
		t.Error(err.Error())
	}

	// Get events channel
	session := sessionBuildResult.Session
	sessionBus := session.GetEventBus()
	sessionChan := sessionBus.GetChan()

	// Assert application builds the session
	events_assertions.AssertApplicationSessionSucceeded(firstApplicationChan, t)

	// Assert session fails to get created
	events_assertions.AssertSessionGetsBuiltAndGetsAvailable(sessionChan, t)

	// Creating the third commit
	thirdCommit := fetcher.NewCommit("Third commit")
	fetcher.AddCommitToBranch(thirdCommit, branch)

	// Fetch the repo again, for the first application, watching started branches
	mediator.ApplicationFetch.Enqueue(firstApplication, true)

	// Assert first application gets fetched SWAPPING the previously built session
	events_assertions.AssertApplicationGetsFetchedWithHotSwap(firstApplicationChan, t)

	// Fetch the repo again, for the second application, watching started branches
	mediator.ApplicationFetch.Enqueue(secondApplication, true)

	// Assert first application gets fetched NOT swapping the session
	// previously built by the other application NOR auto-starting a new session
	// because there were no applications previously started by this application
	events_assertions.AssertApplicationGetsFetched(secondApplicationChan, t)
}
