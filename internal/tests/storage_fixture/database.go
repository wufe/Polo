package storage_fixture

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/wufe/polo/pkg/logging"
)

type FixtureDatabase struct {
	db *badger.DB
}

func NewDB(log logging.Logger) *FixtureDatabase {
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
