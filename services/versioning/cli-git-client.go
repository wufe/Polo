package versioning

import (
	"os/exec"
)

type CLIGitClient struct{}

func NewCLIGitClient() GitClient {
	return &CLIGitClient{}
}

func (client *CLIGitClient) Clone(baseFolder string, outputFolder string, remote string) error {
	cmd := exec.Command("git", "clone", remote, outputFolder)
	cmd.Dir = baseFolder
	err := cmd.Run()
	return err
}

func (client *CLIGitClient) FetchAll(repoFolder string) error {
	cmd := exec.Command("git", "fetch", "--force", "-u", "origin", "+refs/*:refs/*", "--prune")
	cmd.Dir = repoFolder
	err := cmd.Run()
	return err
}
