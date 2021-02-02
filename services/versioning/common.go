package versioning

import (
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/wufe/polo/models"
)

type GitClient interface {
	Clone(baseFolder string, outputFolder string, remote string) error
	FetchAll(repoFolder string) error
}

func GetGitClient(service *models.Service, auth transport.AuthMethod) GitClient {
	if service.UseGitCLI {
		return NewCLIGitClient()
	}
	return NewEmbeddedGitClient(auth)
}
