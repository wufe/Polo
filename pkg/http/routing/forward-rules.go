package routing

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/models"
)

type ForwardRules func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request)

func BuildDefaultForwardRules(conf *models.ApplicationConfiguration, variables models.Variables) (ForwardRules, error) {

	target := conf.Target
	target = variables.ApplyTo(target)

	url, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	return (func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
		r.URL.Host = url.Host
		r.URL.Scheme = url.Scheme
		r.Host = url.Host

		if conf.Host != "" {
			r.Header.Add("Host", conf.Host)
			r.Host = conf.Host
		}

		err := conf.Headers.ApplyTo(r)
		if err != nil {
			log.Errorf("Error applying headers to the request: %s", err.Error())
		}

		return w, r
	}), nil
}

func BuildForwardRules(r *http.Request, pattern models.CompiledForwardPattern, conf *models.ApplicationConfiguration, variables models.Variables) (ForwardRules, error) {
	defaultTarget := conf.Target
	defaultTarget = variables.ApplyTo(defaultTarget)
	defaultTo, err := url.Parse(defaultTarget)
	if err != nil {
		return nil, err
	}

	path := r.URL.Path
	if r.URL.RawQuery != "" {
		path += "?" + r.URL.RawQuery
	}

	matches := pattern.Pattern.FindStringSubmatch(path)
	log.WithField("matches", matches).Traceln("Matching additional forward rule")

	target := pattern.Forward.To
	for index, match := range matches {
		target = strings.ReplaceAll(target, fmt.Sprintf("$%d", index), match)
	}
	target = variables.ApplyTo(target)

	to, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	return func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {

		if to.IsAbs() {
			r.URL = to
			r.Host = to.Host
		} else {
			r.URL = defaultTo
			r.URL.Host = defaultTo.Host
			r.URL.Scheme = defaultTo.Scheme
			r.Host = defaultTo.Host
			r.URL.Path = to.Path
			r.URL.RawQuery = to.RawQuery
		}

		if pattern.Forward.Host != "" {
			r.Header.Add("Host", pattern.Forward.Host)
			r.Host = pattern.Forward.Host
		}

		err := pattern.Forward.Headers.ApplyTo(r)
		if err != nil {
			log.Errorf("Error applying headers to the request: %s", err.Error())
		}

		return w, r
	}, nil
}
