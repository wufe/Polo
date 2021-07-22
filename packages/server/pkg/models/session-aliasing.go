package models

import (
	"github.com/google/uuid"
	"github.com/wufe/polo/pkg/namesgenerator"
)

func NewSessionAlias(exclude []string) string {
	genCount := 0
	for {
		retry := 0
		if genCount > 50 {
			retry = 1
		}
		name := namesgenerator.GetRandomName(retry)
		found := false
		for _, e := range exclude {
			if e == name {
				found = true
			}
		}
		if !found {
			return name
		}
		genCount++
		if genCount > 500 {
			break
		}
	}
	return uuid.NewString() + "-alias"
}
