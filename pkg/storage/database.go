package storage

import (
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
	"github.com/wufe/polo/pkg/utils"
)

func StartDB() (*badger.DB, error) {

	exeFolder := utils.GetExecutableFolder()

	db, err := badger.Open(badger.DefaultOptions(filepath.Join(exeFolder, "./db")))
	if err != nil {
		return nil, err
	}
	return db, nil
}
