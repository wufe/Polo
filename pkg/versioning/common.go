package versioning

type GitClient interface {
	Clone(baseFolder string, outputFolder string, remote string) error
	FetchAll(repoFolder string) error
	HardReset(repoFolder string, commit string) error
}

func GetGitClient() GitClient {
	// Using CLI only
	return NewCLIGitClient()
}
