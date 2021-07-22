package background

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/wufe/polo/pkg/execution"
	"github.com/wufe/polo/pkg/http/net"
	"github.com/wufe/polo/pkg/models"
)

type SessionCommandExecution interface {
	ExecCommand(ctx context.Context, command *models.Command, session *models.Session) error
}

type sessionCommandExecutionImpl struct {
	portRetriever net.PortRetriever
	commandRunner execution.CommandRunner
}

func NewSessionCommandExecution(portRetriever net.PortRetriever, commandRunner execution.CommandRunner) SessionCommandExecution {
	return &sessionCommandExecutionImpl{
		portRetriever: portRetriever,
		commandRunner: commandRunner,
	}
}

func (ce *sessionCommandExecutionImpl) ExecCommand(ctx context.Context, command *models.Command, session *models.Session) error {
	builtCommand, err := ce.buildCommand(command.Command, session)
	if err != nil {
		return err
	}
	session.LogStdin(builtCommand)

	cmdCtx := ctx
	if command.Timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(command.Timeout)*time.Second)
		defer cancel()
		cmdCtx = timeoutCtx
	}
	cmds := ParseCommandContext(cmdCtx, builtCommand)
	for _, cmd := range cmds {
		cmd.Env = append(
			os.Environ(),
			command.Environment...,
		)
		cmd.Dir = getWorkingDir(session.Folder, command.WorkingDir)
	}

	err = ce.commandRunner.ExecCmds(ctx, func(line *execution.StdLine) {
		if line.Type == execution.StdTypeOut {
			session.LogStdout(line.Line)
		} else {
			session.LogStderr(line.Line)
		}
		parseSessionCommandOuput(session, command, line.Line)
	}, cmds...)

	return err
}

func (ce *sessionCommandExecutionImpl) buildCommand(command string, session *models.Session) (string, error) {
	ce.addPortsOnDemand(command, session)
	command = session.Variables.ApplyTo(command)
	return strings.TrimSpace(command), nil
}

func (ce *sessionCommandExecutionImpl) addPortsOnDemand(input string, session *models.Session) (string, error) {
	conf := session.GetConfiguration()
	re := regexp.MustCompile(`{{(port\d*)}}`)
	matches := re.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		portVariable := match[1]
		if _, ok := session.Variables[portVariable]; !ok {
			port, err := ce.portRetriever.GetFreePort(conf.Port)
			if err != nil {
				return "", err
			}
			session.SetVariable(portVariable, fmt.Sprint(port))
		}
	}
	return input, nil
}

func ParseCommandContext(context context.Context, command string) []*exec.Cmd {

	commands := []*exec.Cmd{}

	for _, name := range strings.Split(command, "|") {
		name = strings.TrimSpace(name)
		nameAndArgs := strings.Split(name, " ")

		if runtime.GOOS == "windows" {
			nameAndArgs = append([]string{"cmd", "/C"}, nameAndArgs...)
		}

		cmd := exec.CommandContext(context, nameAndArgs[0], nameAndArgs[1:]...)
		commands = append(commands, cmd)
	}

	return commands
}
