package session_healthcheck

import (
	"testing"
	"time"

	"github.com/wufe/polo/internal/tests"
	"github.com/wufe/polo/internal/tests/events_assertions"
	"github.com/wufe/polo/internal/tests/execution_fixture"
	"github.com/wufe/polo/internal/tests/net_fixture"
	"github.com/wufe/polo/internal/tests/versioning_fixture"
	"github.com/wufe/polo/pkg/models"
)

// Session should become available using a mock http server
func Test_SessionShouldBuildAndBeAvailableWithFixtureHTTPServer(t *testing.T) {

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
	}, &models.ApplicationConfiguration{
		SharedConfiguration: models.SharedConfiguration{
			Remote: "FakeRemote",
			Commands: models.Commands{
				Start: []models.Command{
					{Command: "valid-command.exe"},
				},
				Stop: []models.Command{
					{Command: "valid-command.exe"},
				},
			},
			Startup: models.Startup{
				Retries: 3,
			},
			Healthcheck: models.Healthcheck{
				RetryInterval: 0.001,
			},
		},
		Name:      "Test_SessionShouldBuildAndBeAvailableWithFixtureHTTPServer",
		IsDefault: true,
		Branches: []models.BranchConfigurationMatch{
			{
				Test: "main",
				BranchConfiguration: models.BranchConfiguration{
					Watch: false,
					Main:  false,
				},
			},
		},
	})

	// Get events channel
	applications := di.GetApplications()
	firstApplication := applications[0]
	firstApplicationBus := firstApplication.GetEventBus()
	firstApplicationChan := firstApplicationBus.GetChan()

	// Assert application is being loaded
	events_assertions.AssertApplicationEvents(
		firstApplicationChan,
		[]models.ApplicationEventType{
			models.ApplicationEventTypeInitializationStarted,
			models.ApplicationEventTypeFetchStarted,
			models.ApplicationEventTypeFetchCompleted,
			models.ApplicationEventTypeInitializationCompleted,
		},
		t,
		10*time.Second,
	)

	// Creating the second commit
	secondCommit := fetcher.NewCommit("Second commit")
	fetcher.AddCommitToBranch(secondCommit, branch)

	// Re-fetch the application
	mediator := di.GetMediator()
	mediator.ApplicationFetch.Enqueue(firstApplication, false)

	// Assert application gets fetched
	events_assertions.AssertApplicationEvents(
		firstApplicationChan,
		[]models.ApplicationEventType{
			models.ApplicationEventTypeFetchStarted,
			models.ApplicationEventTypeFetchCompleted,
		},
		t,
		2*time.Second,
	)

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
		},
		t,
		10*time.Second,
	)
}
