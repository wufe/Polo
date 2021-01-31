package utils

import "os"

func IsDev() bool {
	return os.Getenv("GO_ENV") == "development"
}

func DevServerURL() string {
	url := os.Getenv("WDS_URL")
	if url == "" {
		url = "http://localhost:9000"
	}
	return url
}
