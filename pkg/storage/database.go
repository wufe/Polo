package storage

import (
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
	log "github.com/sirupsen/logrus"
)

type Database struct {
	DB *badger.DB
}

func NewDB(folder string) *Database {
	options := badger.DefaultOptions(filepath.Join(folder, "./db"))
	options.Logger = nil
	db, err := badger.Open(options)
	if err != nil {
		log.Panicf("Error while opening database: %s", err.Error())
	}
	return &Database{
		DB: db,
	}
}
