package session_build

import (
	"testing"
	"time"

	"github.com/wufe/polo/internal/tests"
	"github.com/wufe/polo/internal/tests/events_assertions"
	"github.com/wufe/polo/internal/tests/versioning_fixture"
	"github.com/wufe/polo/pkg/models"
)

// The session gets build
// the build process fails
// the clean process starts
// all command fails
// 		since it is set "ContinueOnError: true", the clean process continues
// the clean process ends with folder deletion
func Test_SessionBuildFailingClean(t *testing.T) {

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
	}, &models.ApplicationConfiguration{
		SharedConfiguration: models.SharedConfiguration{
			Remote: "FakeRemote",
			Commands: models.Commands{
				Start: []models.Command{
					{Command: "notexistingcommand.exe"},
				},
				Stop: []models.Command{
					{Command: "notexistingcommand.exe"},
				},
				Clean: []models.Command{
					{Command: "1st_cleancommand.sh", ContinueOnError: true},
					{Command: "2nd_cleancommand.sh", ContinueOnError: true},
					{Command: "3rd_cleancommand.sh", ContinueOnError: true},
				},
			},
		},
		Name:      "Test_SessionBuildFailing",
		IsDefault: true,
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
		10*time.Second,
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
			models.SessionEventTypeBuildStarted,
			models.SessionEventTypePreparingFolders,
			models.SessionEventTypeCommandsExecutionStarted,
			models.SessionEventTypeCommandsExecutionFailed,
			models.SessionEventTypeCleanCommandExecution,
			models.SessionEventTypeCleanCommandExecution,
			models.SessionEventTypeCleanCommandExecution,
			models.SessionEventTypeFolderClean,
		},
		t,
		10*time.Second,
	)
}
