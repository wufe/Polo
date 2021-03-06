package models

import "github.com/wufe/polo/pkg/models/output"

func MapApplication(model *Application) *output.Application {
	if model == nil {
		return nil
	}
	conf := model.GetConfiguration()
	model.RLock()
	defer model.RUnlock()
	return &output.Application{
		Status:        string(model.Status),
		Configuration: MapApplicationConfiguration(conf),
		Folder:        model.Folder,
		BaseFolder:    model.BaseFolder,
		BranchesMap:   MapBranches(model.BranchesMap),
	}
}

func MapApplicationConfiguration(model ApplicationConfiguration) output.ApplicationConfiguration {
	return output.ApplicationConfiguration{
		Name:                  model.Name,
		Remote:                model.Remote,
		Target:                model.Target,
		Host:                  model.Host,
		Fetch:                 MapFetch(model.Fetch),
		Helper:                MapHelper(model.Helper),
		IsDefault:             model.IsDefault,
		Forwards:              MapForwards(model.Forwards),
		Headers:               MapHeaders(model.Headers),
		Healthcheck:           MapHealthcheck(model.Healthcheck),
		Startup:               MapStartup(model.Startup),
		Recycle:               MapRecycle(model.Recycle),
		Commands:              MapCommands(model.Commands),
		MaxConcurrentSessions: model.MaxConcurrentSessions,
		Port:                  MapPort(model.Port),
		UseFolderCopy:         model.UseFolderCopy,
		CleanOnExit:           *model.CleanOnExit,
		Warmup:                MapWarmups(model.Warmup),
	}
}

func MapWarmups(model Warmups) output.Warmups {
	urls := []output.Warmup{}
	for _, u := range model.URLs {
		urls = append(urls, output.Warmup{
			Method:  u.Method,
			URL:     u.URL,
			Status:  u.Status,
			Timeout: u.Timeout,
		})
	}
	return output.Warmups{
		MaxRetries:    model.MaxRetries,
		RetryInterval: model.RetryInterval,
		URLs:          urls,
	}
}

func MapApplications(models []*Application) []output.Application {
	ret := []output.Application{}
	for _, a := range models {
		ret = append(ret, *MapApplication(a))
	}
	return ret
}

func MapFetch(model Fetch) output.Fetch {
	return output.Fetch{
		Interval: model.Interval,
	}
}

func MapHelper(model Helper) output.Helper {
	return output.Helper{
		Position: string(model.Position),
	}
}

func MapForward(model Forward) output.Forward {
	return output.Forward{
		Pattern: model.Pattern,
		To:      model.To,
		Host:    model.Host,
		Headers: MapHeaders(model.Headers),
	}
}

func MapForwards(models []Forward) []output.Forward {
	ret := []output.Forward{}
	for _, f := range models {
		ret = append(ret, MapForward(f))
	}
	return ret
}

func MapHeaders(model Headers) output.Headers {
	add := []string{}
	set := []string{}
	replace := []string{}
	for _, a := range model.Add {
		add = append(add, string(a))
	}
	for _, s := range model.Set {
		set = append(set, string(s))
	}
	for _, r := range model.Replace {
		replace = append(replace, string(r))
	}
	return output.Headers{
		Add:     add,
		Set:     set,
		Del:     model.Del,
		Replace: replace,
	}
}

func MapHealthcheck(model Healthcheck) output.Healthcheck {
	return output.Healthcheck{
		Method:        model.Method,
		URL:           model.URL,
		Status:        model.Status,
		MaxRetries:    model.MaxRetries,
		RetryInterval: model.RetryInterval,
		Timeout:       model.Timeout,
	}
}

func MapStartup(model Startup) output.Startup {
	return output.Startup{
		Timeout: model.Timeout,
		Retries: model.Retries,
	}
}

func MapRecycle(model Recycle) output.Recycle {
	return output.Recycle{
		InactivityTimeout: model.InactivityTimeout,
	}
}

func MapCommand(model Command) output.Command {
	return output.Command{
		Command:             model.Command,
		Environment:         model.Environment,
		OutputVariable:      model.OutputVariable,
		ContinueOnError:     model.ContinueOnError,
		WorkingDir:          model.WorkingDir,
		StartHealthchecking: model.StartHealthchecking,
		Timeout:             model.Timeout,
	}
}

func MapCommands(model Commands) output.Commands {
	start := []output.Command{}
	stop := []output.Command{}
	for _, s := range model.Start {
		start = append(start, MapCommand(s))
	}
	for _, s := range model.Stop {
		stop = append(stop, MapCommand(s))
	}
	return output.Commands{
		Start: start,
		Stop:  stop,
	}
}

func MapPort(model PortConfiguration) output.PortConfiguration {
	return output.PortConfiguration{
		Except: model.Except,
	}
}

func MapBranch(model Branch) output.Branch {
	return output.Branch{
		Name:    model.Name,
		Hash:    model.Hash,
		Author:  model.Author,
		Date:    model.Date,
		Message: model.Message,
	}
}

func MapBranches(model map[string]*Branch) map[string]output.Branch {
	ret := make(map[string]output.Branch)
	for k, v := range model {
		ret[k] = MapBranch(*v)
	}
	return ret
}
