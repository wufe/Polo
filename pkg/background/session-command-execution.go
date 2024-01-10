package background

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/creack/pty"
	"github.com/wufe/polo/pkg/logging"

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
	configuration *models.RootConfiguration
	log           logging.Logger
}

func NewSessionCommandExecution(
	portRetriever net.PortRetriever,
	commandRunner execution.CommandRunner,
	configuration *models.RootConfiguration,
	log logging.Logger,
) SessionCommandExecution {
	return &sessionCommandExecutionImpl{
		portRetriever: portRetriever,
		commandRunner: commandRunner,
		configuration: configuration,
		log:           log,
	}
}

func (ce *sessionCommandExecutionImpl) ExecCommand(ctx context.Context, command *models.Command, session *models.Session) error {
	builtCommand, err := ce.buildCommand(command.Command, session)
	if err != nil {
		return err
	}

	cmdCtx := ctx
	if command.Timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(command.Timeout)*time.Second)
		defer cancel()
		cmdCtx = timeoutCtx
	}

	cmd := ParseCommandContext(cmdCtx, builtCommand)
	cmd.Env = append(
		os.Environ(),
		ce.buildEnvironmentVariables(command.Environment, session)...,
	)
	cmd.Dir = getWorkingDir(session.Folder, command.WorkingDir)

	session.LogStdin([]byte(builtCommand))

	if ce.configuration.Global.FeaturesPreview.AdvancedTerminalOutput {
		tty, err := pty.Start(cmd)
		if err != nil {
			return fmt.Errorf("error starting pty: %w", err)
		}

		outputBuffer := session.GetTTYOutput()

		buffer := make([]byte, 1024)

		for {
			n, err := tty.Read(buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				ce.log.Warnf("Error reading from PTY: %s", err)
				break
			}

			if _, err := outputBuffer.Write(buffer[:n]); err != nil {
				ce.log.Warnf("Error writing to PTY output buffer: %s", err)
			}
		}
	} else {
		err = ce.commandRunner.ExecCmds(ctx, func(line *execution.StdLine) {
			if line.Type == execution.StdTypeOut {
				session.LogStdout([]byte(line.Line))
			} else {
				session.LogStderr([]byte(line.Line))
			}
			parseSessionCommandOuput(session, command, line.Line)
		}, cmd)
	}

	return err
}

func (ce *sessionCommandExecutionImpl) buildEnvironmentVariables(variables []string, session *models.Session) []string {
	ret := []string{}
	for _, variable := range variables {
		ret = append(ret, session.Variables.ApplyTo(variable))
	}
	return ret
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

func ParseCommandContext(context context.Context, command string) *exec.Cmd {
	var rawCmd []string

	if runtime.GOOS == "windows" {
		rawCmd = []string{"cmd", "/S", "/C"}
	} else {
		rawCmd = []string{"/bin/sh", "-c"}
	}

	rawCmd = append(rawCmd, command)

	return exec.CommandContext(context, rawCmd[0], rawCmd[1:]...)
}
