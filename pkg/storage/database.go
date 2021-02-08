package storage

import (
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
	log "github.com/sirupsen/logrus"
	"github.com/wufe/polo/pkg/utils"
)

type Database struct {
	DB *badger.DB
}

func NewDB() *Database {
	exeFolder := utils.GetExecutableFolder()

	options := badger.DefaultOptions(filepath.Join(exeFolder, "./db"))
	options.Logger = nil
	db, err := badger.Open(options)
	if err != nil {
		log.Panicf("Error while opening database: %s", err.Error())
	}
	return &Database{
		DB: db,
	}
}
