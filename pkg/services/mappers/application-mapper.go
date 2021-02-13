package mappers

import (
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/services/output"
)

func MapApplication(model *models.Application) *output.Application {
	if model == nil {
		return nil
	}
	model.Lock()
	defer model.Unlock()
	return &output.Application{
		Status:                string(model.Status),
		Name:                  model.Name,
		Remote:                model.Remote,
		Target:                model.Target,
		Host:                  model.Host,
		Fetch:                 MapFetch(model.Fetch),
		Watch:                 model.Watch,
		IsDefault:             model.IsDefault,
		Forwards:              MapForwards(model.Forwards),
		Headers:               MapHeaders(model.Headers),
		Healthcheck:           MapHealthcheck(model.Healthcheck),
		Startup:               MapStartup(model.Startup),
		Recycle:               MapRecycle(model.Recycle),
		Commands:              MapCommands(model.Commands),
		MaxConcurrentSessions: model.MaxConcurrentSessions,
		Port:                  MapPort(model.Port),
		UseGitCLI:             model.UseGitCLI,
		Folder:                model.Folder,
		BaseFolder:            model.BaseFolder,
		Branches:              MapBranches(model.Branches),
	}
}

func MapApplications(models []*models.Application) []output.Application {
	ret := []output.Application{}
	for _, a := range models {
		ret = append(ret, *MapApplication(a))
	}
	return ret
}

func MapFetch(model models.Fetch) output.Fetch {
	return output.Fetch{
		Interval: model.Interval,
	}
}

func MapForward(model models.Forward) output.Forward {
	return output.Forward{
		Pattern: model.Pattern,
		To:      model.To,
		Host:    model.Host,
		Headers: MapHeaders(model.Headers),
	}
}

func MapForwards(models []models.Forward) []output.Forward {
	ret := []output.Forward{}
	for _, f := range models {
		ret = append(ret, MapForward(f))
	}
	return ret
}

func MapHeaders(model models.Headers) output.Headers {
	add := []string{}
	set := []string{}
	for _, a := range model.Add {
		add = append(add, string(a))
	}
	for _, s := range model.Set {
		set = append(set, string(s))
	}
	return output.Headers{
		Add: add,
		Set: set,
		Del: model.Del,
	}
}

func MapHealthcheck(model models.Healthcheck) output.Healthcheck {
	return output.Healthcheck{
		Method:        model.Method,
		URL:           model.URL,
		Status:        model.Status,
		MaxRetries:    model.MaxRetries,
		RetryInterval: model.RetryInterval,
		RetryTimeout:  model.RetryTimeout,
	}
}

func MapStartup(model models.Startup) output.Startup {
	return output.Startup{
		Timeout: model.Timeout,
		Retries: model.Retries,
	}
}

func MapRecycle(model models.Recycle) output.Recycle {
	return output.Recycle{
		InactivityTimeout: model.InactivityTimeout,
	}
}

func MapCommand(model models.Command) output.Command {
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

func MapCommands(model models.Commands) output.Commands {
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

func MapPort(model models.PortConfiguration) output.PortConfiguration {
	return output.PortConfiguration{
		Except: model.Except,
	}
}

func MapBranch(model models.Branch) output.Branch {
	return output.Branch{
		Name:    model.Name,
		Hash:    model.Hash,
		Author:  model.Author,
		Date:    model.Date,
		Message: model.Message,
	}
}

func MapBranches(model map[string]*models.Branch) map[string]output.Branch {
	ret := make(map[string]output.Branch)
	for k, v := range model {
		ret[k] = MapBranch(*v)
	}
	return ret
}
