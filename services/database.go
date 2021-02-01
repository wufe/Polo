package services

import (
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
)

func StartDB() (*badger.DB, error) {

	exeFolder := getExecutableFolder()

	db, err := badger.Open(badger.DefaultOptions(filepath.Join(exeFolder, "./db")))
	if err != nil {
		return nil, err
	}
	return db, nil
}
