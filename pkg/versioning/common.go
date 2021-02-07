package versioning

import (
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/wufe/polo/pkg/models"
)

type GitClient interface {
	Clone(baseFolder string, outputFolder string, remote string) error
	FetchAll(repoFolder string) error
}

func GetGitClient(application *models.Application, auth transport.AuthMethod) GitClient {
	if application.UseGitCLI {
		return NewCLIGitClient()
	}
	return NewEmbeddedGitClient(auth)
}
