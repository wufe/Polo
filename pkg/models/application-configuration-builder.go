package models

func BuildApplicationConfiguration(name string) *ApplicationConfiguration {
	return &ApplicationConfiguration{
		SharedConfiguration: SharedConfiguration{
			Remote: "",
			Commands: Commands{
				Start: []Command{},
				Stop:  []Command{},
			},
			Startup:     Startup{},
			Healthcheck: Healthcheck{},
		},
		Name:      name,
		IsDefault: false,
		Branches:  []BranchConfigurationMatch{},
	}
}

func (a *ApplicationConfiguration) WithRemote(remote string) *ApplicationConfiguration {
	a.Remote = remote
	return a
}

func (a *ApplicationConfiguration) WithStartCommand(command string) *ApplicationConfiguration {
	a.Commands.Start = append(a.Commands.Start, Command{Command: command})
	return a
}

func (a *ApplicationConfiguration) WithStopCommand(command string) *ApplicationConfiguration {
	a.Commands.Stop = append(a.Commands.Stop, Command{Command: command})
	return a
}

func (a *ApplicationConfiguration) WithStartupRetries(n int) *ApplicationConfiguration {
	a.Startup.Retries = n
	return a
}

func (a *ApplicationConfiguration) WithHealthcheckRetryInterval(interval float32) *ApplicationConfiguration {
	a.Healthcheck.RetryInterval = interval
	return a
}

func (a *ApplicationConfiguration) SetAsDefault(def bool) *ApplicationConfiguration {
	a.IsDefault = def
	return a
}

func (a *ApplicationConfiguration) WithBranch(branch *BranchConfigurationMatch) *ApplicationConfiguration {
	a.Branches = append(a.Branches, *branch)
	return a
}
