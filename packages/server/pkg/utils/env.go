package utils

import (
	"log"
	"os"
	"path/filepath"
)

type Environment interface {
	IsTest() bool
	IsDev() bool
	IsDebugRace() bool
	IsDiagnostics() bool
	DevServerURL() string
	GetExecutableFolder() string
}

type environmentImpl struct{}

func DetectEnvironment() Environment {
	return &environmentImpl{}
}

func (e *environmentImpl) IsTest() bool {
	return false
}

func (e *environmentImpl) IsDiagnostics() bool {
	return os.Getenv("POLO_DIAGNOSTICS") == "true"
}

func (e *environmentImpl) IsDev() bool {
	return os.Getenv("GO_ENV") == "development"
}

func (e *environmentImpl) IsDebugRace() bool {
	return os.Getenv("GO_DEBUG") == "race"
}

func (e *environmentImpl) DevServerURL() string {
	url := os.Getenv("WDS_URL")
	if url == "" {
		url = "http://localhost:9000"
	}
	return url
}

func (e *environmentImpl) GetExecutableFolder() string {
	if cwd := os.Getenv("POLO_CWD"); cwd != "" {
		return cwd
	}
	if e.IsDev() {
		path, err := os.Getwd()
		if err != nil {
			log.Fatalln("Error retrieving file path", err)
		}
		return path
	}
	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalln("Error retrieving file path", err)
	}
	dir := filepath.Dir(executablePath)
	return dir
}
