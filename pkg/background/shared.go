package background

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/phayes/freeport"
	"github.com/wufe/polo/pkg/models"
)

func getWorkingDir(baseDir string, commandWorkingDir string) string {
	if commandWorkingDir == "" {
		return baseDir
	}
	if strings.HasPrefix(commandWorkingDir, "./") || !strings.HasPrefix(commandWorkingDir, "/") {
		return filepath.Join(baseDir, commandWorkingDir)
	}
	return commandWorkingDir
}

func parseSessionCommandOuput(session *models.Session, command *models.Command, output string) {
	session.SetVariable("last_output", output)
	re := regexp.MustCompile(`polo\[([^\]]+?)=([^\]]+?)\]`)
	matches := re.FindAllStringSubmatch(output, -1)
	for _, variable := range matches {
		key := variable[1]
		value := variable[2]
		session.SetVariable(key, value)
		session.LogWarn(fmt.Sprintf("Setting variable %s=%s", key, value))
	}

	if command.OutputVariable != "" {
		session.SetVariable(command.OutputVariable, output)
	}
}

func buildCommand(command string, session *models.Session) (string, error) {
	addPortsOnDemand(command, session)
	command = session.Variables.ApplyTo(command)
	return strings.TrimSpace(command), nil
}

func addPortsOnDemand(input string, session *models.Session) (string, error) {
	re := regexp.MustCompile(`{{(port\d*)}}`)
	matches := re.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		portVariable := match[1]
		if _, ok := session.Variables[portVariable]; !ok {
			port, err := getFreePort(&session.Application.Port)
			if err != nil {
				return "", err
			}
			session.SetVariable(portVariable, fmt.Sprint(port))
		}
	}
	return input, nil
}

func getFreePort(portConfiguration *models.PortConfiguration) (int, error) {
	foundPort := 0
L:
	for foundPort == 0 {
		freePort, err := freeport.GetFreePort()
		if err != nil {
			return 0, err
		}
		for _, port := range portConfiguration.Except {
			if freePort == port {
				continue L
			}
		}
		foundPort = freePort
	}
	return foundPort, nil
}
