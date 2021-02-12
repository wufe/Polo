package utils

import (
	"log"
	"os"
	"path/filepath"
)

func IsDev() bool {
	return os.Getenv("GO_ENV") == "development"
}

func IsDebugRace() bool {
	return os.Getenv("GO_DEBUG") == "race"
}

func DevServerURL() string {
	url := os.Getenv("WDS_URL")
	if url == "" {
		url = "http://localhost:9000"
	}
	return url
}

func GetExecutableFolder() string {
	if IsDev() {
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
