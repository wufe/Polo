package utils

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	pipes "github.com/ebuchman/go-shell-pipes"
	log "github.com/sirupsen/logrus"
)

func ExecuteCommand(cmds ...*exec.Cmd) (outputChan chan string, doneChan chan struct{}, err error) {
	if len(cmds) > 1 {
		for i, c := range cmds {
			if i < len(cmds)-1 {
				cmds[i+1].Stdin, _ = c.StdoutPipe()
			}
		}
	}

	lastCmd := cmds[len(cmds)-1]

	errPipe, err := lastCmd.StderrPipe()
	if err != nil {
		log.Errorf("[CMD:%s] %s", lastCmd.Path, err.Error())
		return nil, nil, err
	}
	outPipe, err := lastCmd.StdoutPipe()
	if err != nil {
		log.Errorf("[CMD:%s] %s", lastCmd.Path, err.Error())
		return nil, nil, err
	}

	returnOutputChan := make(chan string)
	returnDoneChan := make(chan struct{})

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		scanner := bufio.NewScanner(errPipe)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) != "" {
				returnOutputChan <- line
			}
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		scanner := bufio.NewScanner(outPipe)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) != "" {
				returnOutputChan <- line
			}
		}
		wg.Done()
	}()

	pipes.RunCmds(cmds)

	go func() {
		wg.Wait()
		lastCmd.Wait()
		close(returnOutputChan)
		returnDoneChan <- struct{}{}
	}()

	return returnOutputChan, returnDoneChan, nil

}

func ThroughCallback(output chan string, done chan struct{}, err error) func(callback func(string)) error {
	return func(callback func(string)) error {
		if err != nil {
			return err
		}

	L:
		for {
			select {
			case line, ok := <-output:
				if ok {
					callback(line)
				}
			case <-done:
				break L
			}
		}

		return nil
	}
}

func ExecCmds(callback func(*StdLine), cmds ...*exec.Cmd) error {

	cmdCtx, cancelCtx := context.WithCancel(context.Background())

	for i := 1; i < len(cmds); i++ {
		cmds[i].Stdin, _ = cmds[i-1].StdoutPipe()
	}

	lastCmd := cmds[len(cmds)-1]

	var wg sync.WaitGroup

	// Start the last
	wg.Add(1)
	var lastCmdErr error
	go func() {
		lastCmdErr = ExecCmd(cmdCtx, lastCmd, callback)
		wg.Done()
	}()

	// Start the others in descending order
	for i := len(cmds) - 2; i >= 0; i-- {
		if err := cmds[i].Start(); err != nil {
			cancelCtx()
			return err
		}
	}

	// Wait for them in ascending order,
	// except for the last
	for i := 0; i < len(cmds)-1; i++ {
		cmds[i].Wait()
	}

	// Wait for the last
	wg.Wait()

	cancelCtx()

	return lastCmdErr
}

func ExecCmd(ctx context.Context, cmd *exec.Cmd, callback func(*StdLine)) error {
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	cmd.Start()

	done := make(chan struct{})
	go func(callback func(*StdLine), done chan struct{}) {
		var wg sync.WaitGroup

		eof := make(chan struct{})
		defer close(eof)
		messages := make(chan *StdLine, 5)
		defer close(messages)

		wg.Add(1)
		go func() {
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				line := scanner.Text()
				messages <- &StdLine{
					Type: StdTypeOut,
					Line: line,
				}
			}
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				line := scanner.Text()
				messages <- &StdLine{
					Type: StdTypeErr,
					Line: line,
				}
			}
			wg.Done()
		}()

		go func() {
			for {
				select {
				case message := <-messages:
					if callback != nil {
						callback(message)
					}
				case <-ctx.Done():
					stdoutPipe.Close()
					stderrPipe.Close()
					eof <- struct{}{}
				case <-eof:
					return
				}
			}
		}()

		wg.Wait()

		eof <- struct{}{}

		done <- struct{}{}
	}(callback, done)
	<-done

	cmd.Wait()

	exitCode := cmd.ProcessState.ExitCode()
	if exitCode > 0 {
		return fmt.Errorf("Command exit with code %d", exitCode)
	}

	return nil
}

type StdLine struct {
	Type StdType
	Line string
}

type StdType string

const StdTypeOut StdType = "stdout"
const StdTypeErr StdType = "stderr"
