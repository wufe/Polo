package utils

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"sync"
)

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
		lastCmdErr = execCmd(cmdCtx, lastCmd, callback)
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

func execCmd(ctx context.Context, cmd *exec.Cmd, callback func(*StdLine)) error {
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
