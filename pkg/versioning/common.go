package versioning

import (
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/wufe/polo/pkg/models"
)

type GitClient interface {
	Clone(baseFolder string, outputFolder string, remote string) error
	FetchAll(repoFolder string) error
	HardReset(repoFolder string, commit string) error
}

func GetGitClient(application *models.Application, auth transport.AuthMethod) GitClient {
	// Using CLI only
	return NewCLIGitClient()
}
