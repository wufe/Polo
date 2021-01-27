package utils

import (
	"bufio"
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
