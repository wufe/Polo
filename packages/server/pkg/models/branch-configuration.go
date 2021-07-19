package models

import (
	"regexp"

	"github.com/wufe/polo/pkg/logging"
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

func (branches Branches) BranchIsBeingWatched(branch string, logger logging.Logger) bool {
	foundBranch, ok := branches.findBranchConfiguration(branch, logger)
	return ok && foundBranch.Watch
}

func (branches Branches) BranchIsMain(branch string, logger logging.Logger) bool {
	foundBranch, ok := branches.findBranchConfiguration(branch, logger)
	return ok && foundBranch.Main
}

func (branches Branches) findBranchConfiguration(name string, logger logging.Logger) (BranchConfigurationMatch, bool) {
	for _, b := range branches {
		test, err := regexp.Compile(b.Test)
		if err != nil {
			logger.Errorf("Could not compile branch test regexp: %s", err.Error())
			continue
		}
		if test.MatchString(name) {
			return b, true
		}
	}
	return BranchConfigurationMatch{}, false
}
