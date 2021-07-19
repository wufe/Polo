package storage

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wufe/polo/pkg/utils"
)

type Instance struct {
	Host string
}

func NewInstance(port string) *Instance {
	var host string
	if port == "443" {
		host = "https://localhost"
	} else {
		host = "http://localhost:" + port
	}
	return &Instance{
		Host: host,
	}
}

func DetectInstance(environment utils.Environment) (*Instance, error) {
	execFolder := environment.GetExecutableFolder()
	hostFilepath := filepath.Join(execFolder, ".host")
	if _, err := os.Stat(hostFilepath); os.IsNotExist(err) {
		return nil, err
	}
	host, err := ioutil.ReadFile(hostFilepath)
	if err != nil {
		return nil, err
	}
	res, err := http.Get(string(host))
	if err != nil {
		return nil, err
	}
	if res.StatusCode == 200 {
		return &Instance{
			Host: string(host),
		}, nil
	}
	return nil, errors.New("Error while retrieving running instance")
}

func (i *Instance) Persist(environment utils.Environment) {
	execFolder := environment.GetExecutableFolder()
	portFilepath := filepath.Join(execFolder, ".host")
	ioutil.WriteFile(portFilepath, []byte(i.Host), os.ModePerm)
}
