package versioning_fixture

import (
	"github.com/wufe/polo/pkg/versioning"
)

type FixtureGitClient struct{}

func NewGitClient() versioning.GitClient {
	return &FixtureGitClient{}
}

func (c *FixtureGitClient) Clone(baseFolder string, outputFolder string, remote string) error {
	// NOOP
	return nil
}

func (c *FixtureGitClient) FetchAll(repoFolder string) error {
	// NOOP
	return nil
}

func (c *FixtureGitClient) HardReset(repoFolder string, commit string) error {
	// NOOP
	return nil
}
