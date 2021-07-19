package events_assertions

import (
	"testing"
	"time"

	"github.com/wufe/polo/pkg/models"
)

func AssertSessionGetsBuiltAndGetsAvailable(sessionChan <-chan models.SessionBuildEvent, t *testing.T) {
	AssertSessionEvents(
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

func AssertSessionFailsToBuild4Times(sessionChan <-chan models.SessionBuildEvent, t *testing.T) {
	AssertSessionEvents(
		sessionChan,
		[]models.SessionEventType{
			// First build
			models.SessionEventTypeBuildStarted,
			models.SessionEventTypePreparingFolders,
			models.SessionEventTypeCommandsExecutionStarted,
			models.SessionEventTypeCommandsExecutionFailed,
			models.SessionEventTypeBuildGettingRetried,

			// First retry
			models.SessionEventTypeBuildStarted,
			models.SessionEventTypePreparingFolders,
			models.SessionEventTypeCommandsExecutionStarted,
			models.SessionEventTypeCommandsExecutionFailed,
			models.SessionEventTypeBuildGettingRetried,

			// Second retry
			models.SessionEventTypeBuildStarted,
			models.SessionEventTypePreparingFolders,
			models.SessionEventTypeCommandsExecutionStarted,
			models.SessionEventTypeCommandsExecutionFailed,
			models.SessionEventTypeBuildGettingRetried,

			// Third retry
			models.SessionEventTypeBuildStarted,
			models.SessionEventTypePreparingFolders,
			models.SessionEventTypeCommandsExecutionStarted,
			models.SessionEventTypeCommandsExecutionFailed,
			models.SessionEventTypeFolderClean,
		},
		t,
		5*time.Second,
	)
}
