package services

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kennygrant/sanitize"
	"github.com/wufe/polo/models"
)

func (sessionHandler *SessionHandler) buildSessionCommitStructure(session *models.Session) (string, error) {
	checkout := sanitize.Name(session.Checkout)
	serviceCommitFolder := filepath.Join(session.Service.ServiceFolder, checkout)
	if _, err := os.Stat(serviceCommitFolder); os.IsNotExist(err) {
		output := sessionHandler.serviceHandler.ExecServiceCommandInServiceFolder(session.Service, &models.ServiceCommand{
			Cmd: *exec.Command("git", "clone", session.Service.Remote, checkout),
		})
		if output.ExitCode > 0 {
			return "", errors.New(fmt.Sprintf("Command exit with code %d", output.ExitCode))
		}
	}
	output := sessionHandler.serviceHandler.ExecServiceCommandInServiceCheckoutFolder(session.Service, &models.ServiceCommand{
		Cmd: *exec.Command("git", "fetch", "--all"),
	}, checkout)
	if output.ExitCode > 0 {
		return "", errors.New(fmt.Sprintf("Command exit with code %d", output.ExitCode))
	}
	output = sessionHandler.serviceHandler.ExecServiceCommandInServiceCheckoutFolder(session.Service, &models.ServiceCommand{
		Cmd: *exec.Command("git", "reset", "--hard", session.Checkout),
	}, checkout)
	if output.ExitCode > 0 {
		return "", errors.New(fmt.Sprintf("Command exit with code %d", output.ExitCode))
	}
	return serviceCommitFolder, nil
}
