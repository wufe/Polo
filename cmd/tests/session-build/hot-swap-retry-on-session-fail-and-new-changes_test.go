package session_build

import (
	"testing"

	"github.com/wufe/polo/internal/tests"
	"github.com/wufe/polo/internal/tests/events_assertions"
	"github.com/wufe/polo/internal/tests/versioning_fixture"
	"github.com/wufe/polo/pkg/models"
)

func Test_HotSwapRetriesOnSessionFailAndNewChanges(t *testing.T) {

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
	}, models.BuildApplicationConfiguration("Test_HotSwapRetriesOnSessionFailAndNewChanges").
		WithRemote("FakeRemote").
		WithStartCommand("notexistingcommand.exe").
		WithStopCommand("notexistingcommand.exe").
		WithStartupRetries(3).
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

	// Get events channel
	session := sessionBuildResult.Session
	sessionBus := session.GetEventBus()
	sessionChan := sessionBus.GetChan()

	// Assert session fails to get created
	events_assertions.AssertSessionFailsToBuild4Times(sessionChan, t)

	// Assert application's sessions get built
	events_assertions.AssertApplicationSessionBuildFails4Times(firstApplicationChan, t)

	// Creating the third commit
	thirdCommit := fetcher.NewCommit("Third commit")
	fetcher.AddCommitToBranch(thirdCommit, branch)

	// Fetch the repo again, watching started branches
	mediator.ApplicationFetch.Enqueue(firstApplication, true)

	// Assert application gets fetched without hot-swap or auto-start
	events_assertions.AssertApplicationGetsFetchedWithHotSwap(firstApplicationChan, t)
}
