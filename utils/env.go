package utils

import "os"

func IsDev() bool {
	return os.Getenv("GO_ENV") == "development"
}
