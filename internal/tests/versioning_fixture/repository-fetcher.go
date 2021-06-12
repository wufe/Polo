package versioning_fixture

import (
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/uuid"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/versioning"
)

type FixtureRepositoryFetcher struct {
	result *versioning.FetchResult
	author object.Signature
}

func NewRepositoryFetcher() *FixtureRepositoryFetcher {
	return &FixtureRepositoryFetcher{
		author: object.Signature{
			Name:  "AutomatedTest",
			Email: "test@test.com",
			When:  time.Now(),
		},
		result: &versioning.FetchResult{
			AppCommits:       []string{},
			ObjectsToHashMap: make(map[string]string),
			HashToObjectsMap: make(map[string]*models.RemoteObject),
			BranchesMap:      make(map[string]*models.Branch),
			TagsMap:          make(map[string]*models.Tag),
			Commits:          []string{},
			CommitMap:        make(map[string]*object.Commit),
		},
	}
}

func (f *FixtureRepositoryFetcher) Fetch(baseFolder string) (*versioning.FetchResult, []*versioning.FetcherError) {
	return f.result, []*versioning.FetcherError{}
}

func (f *FixtureRepositoryFetcher) NewCommit(message string) *object.Commit {
	commit := &object.Commit{
		Hash:      plumbing.NewHash(uuid.NewString()),
		Author:    f.author,
		Committer: f.author,
		Message:   message,
	}
	return commit
}

func (f *FixtureRepositoryFetcher) NewBranch(name string) *models.Branch {
	return &models.Branch{
		CheckoutObject: models.CheckoutObject{
			Name:        name,
			Hash:        "",
			Author:      f.author.Name,
			AuthorEmail: f.author.Email,
			Date:        time.Now(),
			Message:     "",
		},
	}
}

func (f *FixtureRepositoryFetcher) AddCommit(commit *object.Commit) {
	hash := commit.Hash.String()
	f.result.AppCommits = append(f.result.AppCommits, hash)
	f.result.ObjectsToHashMap[hash] = hash
	f.result.Commits = append(f.result.Commits, hash)
	f.result.CommitMap[hash] = commit
}

func (f *FixtureRepositoryFetcher) AddCommitToBranch(commit *object.Commit, branch *models.Branch) {
	hash := commit.Hash.String()
	branch.Hash = hash
	f.AddCommit(commit)
	f.result.ObjectsToHashMap[branch.Name] = hash
	if _, exists := f.result.HashToObjectsMap[hash]; !exists {
		f.result.HashToObjectsMap[hash] = &models.RemoteObject{
			Branches: []string{},
			Tags:     []string{},
		}
	}
	var foundBranch bool
	for _, b := range f.result.HashToObjectsMap[hash].Branches {
		if b == branch.Name {
			foundBranch = true
		}
	}
	if !foundBranch {
		f.result.HashToObjectsMap[hash].Branches = append(f.result.HashToObjectsMap[hash].Branches, branch.Name)
	}
	f.result.BranchesMap[hash] = branch
}
