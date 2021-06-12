package models

import (
	"path"

	"github.com/wufe/polo/pkg/models/output"
)

func mapApplication(model *Application) *output.Application {
	if model == nil {
		return nil
	}
	conf := model.GetConfiguration()
	model.RLock()
	defer model.RUnlock()
	return &output.Application{
		Status:        string(model.Status),
		Filename:      path.Base(model.Filename),
		Configuration: mapApplicationConfiguration(conf),
		Folder:        model.Folder,
		BaseFolder:    model.BaseFolder,
		BranchesMap:   mapBranches(model.BranchesMap),
		TagsMap:       mapTags(model.TagsMap),
		Notifications: mapApplicationNotifications(model.notifications),
	}
}

func mapApplicationConfiguration(model ApplicationConfiguration) output.ApplicationConfiguration {
	return output.ApplicationConfiguration{
		Name:                  model.Name,
		Hash:                  model.Hash,
		ID:                    model.ID,
		Remote:                model.Remote,
		Target:                model.Target,
		Host:                  model.Host,
		Fetch:                 mapFetch(model.Fetch),
		Helper:                mapHelper(model.Helper),
		IsDefault:             model.IsDefault,
		Forwards:              mapForwards(model.Forwards),
		Headers:               mapHeaders(model.Headers),
		Healthcheck:           mapHealthcheck(model.Healthcheck),
		Startup:               mapStartup(model.Startup),
		Recycle:               mapRecycle(model.Recycle),
		Commands:              mapCommands(model.Commands),
		MaxConcurrentSessions: model.MaxConcurrentSessions,
		Port:                  mapPort(model.Port),
		UseFolderCopy:         model.UseFolderCopy,
		CleanOnExit:           *model.CleanOnExit,
		Warmup:                mapWarmups(model.Warmup),
	}
}

func mapWarmups(model Warmups) output.Warmups {
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

// MapApplications converts an application model to an output model
func MapApplications(models []*Application) []output.Application {
	ret := []output.Application{}
	for _, a := range models {
		ret = append(ret, a.ToOutput())
	}
	return ret
}

func mapFetch(model Fetch) output.Fetch {
	return output.Fetch{
		Interval: model.Interval,
	}
}

func mapHelper(model Helper) output.Helper {
	return output.Helper{
		Position: string(model.Position),
	}
}

func MapForward(model Forward) output.Forward {
	return output.Forward{
		Pattern: model.Pattern,
		To:      model.To,
		Host:    model.Host,
		Headers: mapHeaders(model.Headers),
	}
}

func mapForwards(models []Forward) []output.Forward {
	ret := []output.Forward{}
	for _, f := range models {
		ret = append(ret, MapForward(f))
	}
	return ret
}

func mapHeaders(model Headers) output.Headers {
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

func mapHealthcheck(model Healthcheck) output.Healthcheck {
	return output.Healthcheck{
		Method:        model.Method,
		URL:           model.URL,
		Status:        model.Status,
		MaxRetries:    model.MaxRetries,
		RetryInterval: model.RetryInterval,
		Timeout:       model.Timeout,
	}
}

func mapStartup(model Startup) output.Startup {
	return output.Startup{
		Timeout: model.Timeout,
		Retries: model.Retries,
	}
}

func mapRecycle(model Recycle) output.Recycle {
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

func mapCommands(model Commands) output.Commands {
	start := []output.Command{}
	stop := []output.Command{}
	clean := []output.Command{}
	for _, s := range model.Start {
		start = append(start, MapCommand(s))
	}
	for _, s := range model.Stop {
		stop = append(stop, MapCommand(s))
	}
	for _, c := range model.Clean {
		clean = append(clean, MapCommand(c))
	}
	return output.Commands{
		Start: start,
		Stop:  stop,
		Clean: clean,
	}
}

func mapPort(model PortConfiguration) output.PortConfiguration {
	return output.PortConfiguration{
		Except: model.Except,
	}
}

func MapCheckoutObject(model CheckoutObject) output.CheckoutObject {
	return output.CheckoutObject{
		Name:        model.Name,
		Hash:        model.Hash,
		Author:      model.Author,
		AuthorEmail: model.AuthorEmail,
		Date:        model.Date,
		Message:     model.Message,
	}
}

func MapBranch(model Branch) output.Branch {
	return output.Branch{
		CheckoutObject: MapCheckoutObject(model.CheckoutObject),
	}
}

func mapBranches(model map[string]*Branch) map[string]output.Branch {
	ret := make(map[string]output.Branch)
	for k, v := range model {
		ret[k] = MapBranch(*v)
	}
	return ret
}

func MapTag(model Tag) output.Tag {
	return output.Tag{
		CheckoutObject: MapCheckoutObject(model.CheckoutObject),
	}
}

func mapTags(model map[string]*Tag) map[string]output.Tag {
	ret := make(map[string]output.Tag)
	for k, v := range model {
		ret[k] = MapTag(*v)
	}
	return ret
}
