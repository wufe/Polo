package storage_fixture

import (
	log "github.com/sirupsen/logrus"

	"github.com/dgraph-io/badger/v3"
)

type FixtureDatabase struct {
	db *badger.DB
}

func NewDB() *FixtureDatabase {
	options := badger.DefaultOptions("")
	options = options.
		WithSyncWrites(false).
		WithInMemory(true)
	db, err := badger.Open(options)
	if err != nil {
		log.Panicf("Error while opening database: %s", err.Error())
	}
	return &FixtureDatabase{
		db: db,
	}
}

func (d *FixtureDatabase) GetDB() *badger.DB {
	return d.db
}
