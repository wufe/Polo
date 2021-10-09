package storage_fixture

import (
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
	"github.com/wufe/polo/pkg/logging"
	storage_models "github.com/wufe/polo/pkg/storage/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type FixtureDatabase struct {
	db   *badger.DB
	gorm *gorm.DB
}

func NewDB(folder string, log logging.Logger) *FixtureDatabase {
	options := badger.DefaultOptions("")
	options = options.
		WithSyncWrites(false).
		WithInMemory(true)
	db, err := badger.Open(options)
	if err != nil {
		log.Panicf("Error while opening database: %s", err.Error())
	}

	sqliteDB, err := gorm.Open(sqlite.Open(filepath.Join(folder, ".db")), &gorm.Config{})
	if err != nil {
		log.Panicf("Error while opening sqlite database: %s", err.Error())
	}
	sqliteDB.AutoMigrate(&storage_models.User{})

	return &FixtureDatabase{
		db: db,
	}
}

func (d *FixtureDatabase) GetDB() *badger.DB {
	return d.db
}

func (d *FixtureDatabase) GetGorm() *gorm.DB {
	return d.gorm
}
