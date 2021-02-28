package models

import (
	"regexp"

	log "github.com/sirupsen/logrus"
)

type BranchConfigurationMatch struct {
	SharedConfiguration `yaml:",inline"`
	BranchConfiguration `yaml:",inline"`
	Test                string `yaml:"test"`
}

type BranchConfiguration struct {
	Main  bool `yaml:"main"`
	Watch bool `yaml:"watch"`
}

type Branches []BranchConfigurationMatch

func (branches Branches) BranchIsBeingWatched(branch string) bool {
	foundBranch, ok := branches.findBranchConfiguration(branch)
	return ok && foundBranch.Watch
}

func (branches Branches) BranchIsMain(branch string) bool {
	foundBranch, ok := branches.findBranchConfiguration(branch)
	return ok && foundBranch.Main
}

func (branches Branches) findBranchConfiguration(name string) (BranchConfigurationMatch, bool) {
	for _, b := range branches {
		test, err := regexp.Compile(b.Test)
		if err != nil {
			log.Errorf("Could not compile branch test regexp: %s", err.Error())
			continue
		}
		if test.MatchString(name) {
			return b, true
		}
	}
	return BranchConfigurationMatch{}, false
}
