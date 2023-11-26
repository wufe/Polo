package models

func BuildBranchConfigurationMatch(branch string) *BranchConfigurationMatch {
	return &BranchConfigurationMatch{
		Test: branch,
	}
}

func (c *BranchConfigurationMatch) SetWatch(watch bool) *BranchConfigurationMatch {
	c.Watch = watch
	return c
}

func (c *BranchConfigurationMatch) SetMain(main bool) *BranchConfigurationMatch {
	c.Main = main
	return c
}
