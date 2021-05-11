package pkg

import (
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

func ConfigureLogging(container DIContainer) {

	environment := container.GetEnvironment()

	if environment.IsDiagnostics() ||
		environment.IsDev() ||
		environment.IsDebugRace() ||
		environment.IsTest() {
		log.SetLevel(log.TraceLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.SetOutput(ioutil.Discard)

	log.AddHook(&writer.Hook{
		Writer: os.Stderr,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
		},
	})
	log.AddHook(&writer.Hook{
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.TraceLevel,
			log.DebugLevel,
			log.InfoLevel,
			log.DebugLevel,
		},
	})
}
