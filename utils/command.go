package utils

import (
	"bufio"
	"os/exec"
	"strings"
	"sync"
)

func ExecuteCommand(cmd *exec.Cmd) (outputChan chan string, doneChan chan struct{}, err error) {
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	if err := cmd.Start(); err != nil {
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

	go func() {
		wg.Wait()
		cmd.Wait()
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
