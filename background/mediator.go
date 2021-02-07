package background

import "github.com/wufe/polo/background/pipe"

type Mediator struct {
	BuildSession      pipe.SessionBuildPipe
	DestroySession    pipe.SessionDestroyPipe
	SessionFileSystem pipe.SessionFilesystemPipe
	CleanSession      pipe.SessionCleanupPipe
	ApplicationInit   pipe.ApplicationInitPipe
	ApplicationFetch  pipe.ApplicationFetchPipe
}

func NewMediator(
	build pipe.SessionBuildPipe,
	destroy pipe.SessionDestroyPipe,
	fs pipe.SessionFilesystemPipe,
	clean pipe.SessionCleanupPipe,
	init pipe.ApplicationInitPipe,
	fetch pipe.ApplicationFetchPipe,
) *Mediator {
	return &Mediator{
		BuildSession:      build,
		DestroySession:    destroy,
		SessionFileSystem: fs,
		CleanSession:      clean,
		ApplicationInit:   init,
		ApplicationFetch:  fetch,
	}
}
