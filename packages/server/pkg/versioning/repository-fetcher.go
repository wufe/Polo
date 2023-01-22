package versioning

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/wufe/polo/pkg/models"
)

type FetcherError struct {
	Error    error
	Critical bool
}

type RepositoryFetcher interface {
	Fetch(baseFolder string, disableTerminalPrompt bool) (*FetchResult, []*FetcherError)
}
type RepositoryFetcherImpl struct {
	gitClient GitClient
}

func NewRepositoryFetcher(gitClient GitClient) *RepositoryFetcherImpl {
	return &RepositoryFetcherImpl{
		gitClient: gitClient,
	}
}

type FetchResult struct {
	AppCommits       []string
	ObjectsToHashMap map[string]string
	HashToObjectsMap map[string]*models.RemoteObject
	BranchesMap      map[string]*models.Branch
	TagsMap          map[string]*models.Tag
	Commits          []string
	CommitMap        map[string]*object.Commit
}

func (fetcher *RepositoryFetcherImpl) Fetch(baseFolder string, disableTerminalPrompt bool) (*FetchResult, []*FetcherError) {
	objectsToHashMap := make(map[string]string)
	hashToObjectsMap := make(map[string]*models.RemoteObject)
	appBranches := make(map[string]*models.Branch)
	appTags := make(map[string]*models.Tag)
	appCommits := []string{}
	appCommitMap := make(map[string]*object.Commit)
	errors := []*FetcherError{}

	checkObjectExists := func(hashToObjectsMap map[string]*models.RemoteObject) func(hash string) {
		return func(hash string) {
			if _, exists := hashToObjectsMap[hash]; !exists {
				hashToObjectsMap[hash] = &models.RemoteObject{
					Branches: []string{},
					Tags:     []string{},
				}
			}
		}
	}(hashToObjectsMap)

	registerHash := fetcher.registerObjectHash(objectsToHashMap)

	// Open repository
	repo, err := git.PlainOpen(baseFolder)
	if err != nil {
		errors = append(errors, &FetcherError{err, false})
		return nil, errors
	}

	// Fetch
	err = fetcher.gitClient.FetchAll(baseFolder, disableTerminalPrompt)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		errors = append(errors, &FetcherError{
			Error:    fmt.Errorf("%s\n\nEnsure your git cli can do a `fetch` inside %s", err.Error(), baseFolder),
			Critical: true,
		})
	}

	// Branches
	branches, err := repo.Branches()
	if err != nil {
		errors = append(errors, &FetcherError{err, false})
	}
	refPrefix := "refs/heads/"
	branches.ForEach(func(ref *plumbing.Reference) error {
		refName := ref.Name().String()
		refHash := ref.Hash().String()

		if !strings.HasPrefix(refName, refPrefix) {
			return nil
		}
		branchName := refName[len(refPrefix):]

		registerHash(branchName, refHash)
		registerHash(fmt.Sprintf("origin/%s", branchName), refHash)
		registerHash(ref.Name().String(), refHash)

		checkObjectExists(refHash)

		hashToObjectsMap[refHash].Branches = appendWithoutDup(hashToObjectsMap[refHash].Branches, branchName)

		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			return err
		}

		appBranches[branchName] = &models.Branch{
			CheckoutObject: models.CheckoutObject{
				Name:        branchName,
				Hash:        refHash,
				Author:      commit.Author.Name,
				AuthorEmail: commit.Author.Email,
				Date:        commit.Author.When,
				Message:     commit.Message,
			},
		}

		return nil
	})
	if err != nil {
		errors = append(errors, &FetcherError{err, false})
	}

	// Tags
	tags, err := repo.Tags()
	if err != nil {
		errors = append(errors, &FetcherError{err, false})
		return nil, errors
	}

	tagPrefix := "refs/tags/"
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		refName := ref.Name().String()
		refHash := ref.Hash().String()

		if !strings.HasPrefix(refName, tagPrefix) {
			return nil
		}

		tagName := refName[len(tagPrefix):]
		registerHash(tagName, refHash)

		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			return err
		}

		// appTags = appendWithoutDup(appTags, tagName)
		appTags[tagName] = &models.Tag{
			CheckoutObject: models.CheckoutObject{
				Name:        tagName,
				Hash:        refHash,
				Author:      commit.Author.Name,
				AuthorEmail: commit.Author.Email,
				Date:        commit.Author.When,
				Message:     commit.Message,
			},
		}
		registerHash(refName, refHash)
		checkObjectExists(refHash)

		hashToObjectsMap[refHash].Tags = appendWithoutDup(hashToObjectsMap[refHash].Tags, tagName)

		return nil
	})

	// Annotated tags
	tagObjects, err := repo.TagObjects()
	if err != nil {
		errors = append(errors, &FetcherError{err, false})
		return nil, errors
	}

	err = tagObjects.ForEach(func(ref *object.Tag) error {
		refName := ref.Name
		refHash := ref.Hash.String()

		tagName := refName

		registerHash(tagName, refHash)

		commit, err := ref.Commit()
		if err != nil {
			return err
		}

		// appTags = appendWithoutDup(appTags, tagName)
		appTags[tagName] = &models.Tag{
			CheckoutObject: models.CheckoutObject{
				Name:        tagName,
				Hash:        refHash,
				Author:      commit.Author.Name,
				AuthorEmail: commit.Author.Email,
				Date:        commit.Author.When,
				Message:     commit.Message,
			},
		}
		registerHash(refName, refHash)
		checkObjectExists(refHash)

		hashToObjectsMap[refHash].Tags = appendWithoutDup(hashToObjectsMap[refHash].Tags, tagName)

		appCommitMap[refHash] = commit

		return nil
	})

	// Log
	// TODO: Configure "since"
	since := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Now().UTC()
	logs, err := repo.Log(&git.LogOptions{All: true, Since: &since, Until: &until, Order: git.LogOrderCommitterTime})
	if err != nil {
		errors = append(errors, &FetcherError{err, false})
		return nil, errors
	}

	err = logs.ForEach(func(commit *object.Commit) error {
		commitHash := commit.Hash.String()
		registerHash(commitHash, commitHash)
		appCommits = append(appCommits, commitHash)
		appCommitMap[commitHash] = commit
		return nil
	})

	return &FetchResult{
		AppCommits:       appCommits,
		ObjectsToHashMap: objectsToHashMap,
		HashToObjectsMap: hashToObjectsMap,
		BranchesMap:      appBranches,
		TagsMap:          appTags,
		Commits:          appCommits,
		CommitMap:        appCommitMap,
	}, errors
}

func (fetcher *RepositoryFetcherImpl) registerObjectHash(objectsToHashMap map[string]string) func(refName string, hash string) {
	return func(refName, hash string) {
		objectsToHashMap[refName] = hash
	}
}

func appendWithoutDup(slice []string, elem ...string) []string {
	for _, currentElem := range elem {
		foundIndex := -1
		for i, sliceElem := range slice {
			if sliceElem == currentElem {
				foundIndex = i
			}
		}
		if foundIndex == -1 {
			slice = append(slice, currentElem)
		}
	}
	return slice
}
