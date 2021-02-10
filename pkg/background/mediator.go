package background

import "github.com/wufe/polo/pkg/background/pipe"

type Mediator struct {
	BuildSession       pipe.SessionBuildPipe
	DestroySession     pipe.SessionDestroyPipe
	SessionFileSystem  pipe.SessionFilesystemPipe
	CleanSession       pipe.SessionCleanupPipe
	StartSession       pipe.SessionStartPipe
	HealthcheckSession pipe.SessionHealthcheckPipe
	ApplicationInit    pipe.ApplicationInitPipe
	ApplicationFetch   pipe.ApplicationFetchPipe
}

func NewMediator(
	build pipe.SessionBuildPipe,
	destroy pipe.SessionDestroyPipe,
	fs pipe.SessionFilesystemPipe,
	clean pipe.SessionCleanupPipe,
	start pipe.SessionStartPipe,
	healthcheck pipe.SessionHealthcheckPipe,
	init pipe.ApplicationInitPipe,
	fetch pipe.ApplicationFetchPipe,
) *Mediator {
	return &Mediator{
		BuildSession:       build,
		DestroySession:     destroy,
		SessionFileSystem:  fs,
		CleanSession:       clean,
		StartSession:       start,
		HealthcheckSession: healthcheck,
		ApplicationInit:    init,
		ApplicationFetch:   fetch,
	}
}