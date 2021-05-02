package storage

import (
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
	log "github.com/sirupsen/logrus"
)

type Database interface {
	GetDB() *badger.DB
}

type DatabaseImpl struct {
	db *badger.DB
}

func NewDB(folder string) *DatabaseImpl {
	options := badger.DefaultOptions(filepath.Join(folder, "./db"))
	options.Logger = nil
	db, err := badger.Open(options)
	if err != nil {
		log.Panicf("Error while opening database: %s", err.Error())
	}
	return &DatabaseImpl{
		db: db,
	}
}

func (d *DatabaseImpl) GetDB() *badger.DB {
	return d.db
}
