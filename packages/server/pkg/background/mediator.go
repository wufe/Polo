package background

import "github.com/wufe/polo/pkg/background/queues"

type Mediator struct {
	BuildSession       queues.SessionBuildQueue
	DestroySession     queues.SessionDestroyQueue
	SessionFileSystem  queues.SessionFilesystemQueue
	CleanSession       queues.SessionCleanupQueue
	StartSession       queues.SessionStartQueue
	HealthcheckSession queues.SessionHealthcheckQueue
	ApplicationInit    queues.ApplicationInitQueue
	ApplicationFetch   queues.ApplicationFetchQueue
}

func NewMediator(
	build queues.SessionBuildQueue,
	destroy queues.SessionDestroyQueue,
	fs queues.SessionFilesystemQueue,
	clean queues.SessionCleanupQueue,
	start queues.SessionStartQueue,
	healthcheck queues.SessionHealthcheckQueue,
	init queues.ApplicationInitQueue,
	fetch queues.ApplicationFetchQueue,
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
