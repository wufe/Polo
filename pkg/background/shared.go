package background

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

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
		session.LogWarn([]byte(fmt.Sprintf("Setting variable %s=%s", key, value)))
	}

	if command.OutputVariable != "" {
		session.SetVariable(command.OutputVariable, output)
	}
}
