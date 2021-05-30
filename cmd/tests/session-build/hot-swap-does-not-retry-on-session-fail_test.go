package session_build

import (
	"testing"

	"github.com/wufe/polo/internal/tests"
	"github.com/wufe/polo/internal/tests/events_assertions"
	"github.com/wufe/polo/internal/tests/execution_fixture"
	"github.com/wufe/polo/internal/tests/net_fixture"
	"github.com/wufe/polo/internal/tests/versioning_fixture"
	"github.com/wufe/polo/pkg/models"
)

func Test_HotSwapDoesNotRetryOnSessionFail(t *testing.T) {
	fetcher := versioning_fixture.NewRepositoryFetcher()

	// Create the HTTP server and start it
	httpServer := net_fixture.NewHTTPServerFixture()
	port, tearDown := httpServer.Setup()
	defer tearDown()

	// Create port retriever to set up HTTP server port
	portRetriever := net_fixture.NewPortRetrieverFixture()
	portRetriever.SetFreePort(port)

	// Creating the branch
	branch := fetcher.NewBranch("main")

	// Creating command runner
	commandRunner := execution_fixture.NewCommandRunnerFixture()

	// Creating the first commit
	firstCommit := fetcher.NewCommit("First commit")
	fetcher.AddCommitToBranch(firstCommit, branch)

	// Setup the application
	di := tests.Fixture(&tests.InjectableServices{
		RepositoryFetcher: fetcher,
		GitClient:         versioning_fixture.NewGitClient(),
		CommandRunner:     commandRunner,
		PortRetriever:     portRetriever,
	}, models.BuildApplicationConfiguration("Test_HotSwapDoesNotRetryOnSessionFail").
		WithRemote("FakeRemote").
		WithStartCommand("working_command.exe").
		WithStopCommand("working_command.exe").
		WithStartupRetries(3).
		WithHealthcheckRetryInterval(1).
		SetAsDefault(true).
		WithBranch(
			models.BuildBranchConfigurationMatch("main").
				SetWatch(true).
				SetMain(true),
		),
	)

	// Get events channel
	applications := di.GetApplications()
	firstApplication := applications[0]
	firstApplicationBus := firstApplication.GetEventBus()
	firstApplicationChan := firstApplicationBus.GetChan()

	// Assert application is being loaded
	events_assertions.AssertApplicationGetsInitializedAndFetched(firstApplicationChan, t)

	// Request new session to be built
	requestService := di.GetRequestService()
	sessionBuildResult, err := requestService.NewSession(branch.Name, firstApplication.GetConfiguration().Name, false)
	if err != nil {
		t.Error(err.Error())
	}

	// Assert the application is being built
	events_assertions.AssertApplicationSessionSucceeded(firstApplicationChan, t)

	// Get events channel
	session := sessionBuildResult.Session
	sessionBus := session.GetEventBus()
	sessionChan := sessionBus.GetChan()

	// Assert session gets created
	events_assertions.AssertSessionGetsBuiltAndGetsAvailable(sessionChan, t)

	// Creating the second commit
	secondCommit := fetcher.NewCommit("Second commit")
	fetcher.AddCommitToBranch(secondCommit, branch)

	// Set the command runner to fail next command
	// + 3 more retries
	commandRunner.FailNextNCommands(4)

	// Re-fetch the application
	mediator := di.GetMediator()
	mediator.ApplicationFetch.Enqueue(firstApplication, true)

	// Assert application gets fetched
	events_assertions.AssertApplicationGetsFetchedWithHotSwapAndFailingBuild(firstApplicationChan, t)

	// Re-fetch the application
	mediator.ApplicationFetch.Enqueue(firstApplication, true)

	// Assert the application does not get hot-swapped again
	events_assertions.AssertApplicationGetsFetched(firstApplicationChan, t)
}
