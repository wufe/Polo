package storage

import (
	"errors"
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
	"github.com/wufe/polo/pkg/logging"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	storage_models "github.com/wufe/polo/pkg/storage/models"
)

type Database interface {
	GetDB() *badger.DB
	GetGorm() *gorm.DB
}

type DatabaseImpl struct {
	db   *badger.DB
	gorm *gorm.DB
}

func NewDB(folder string, log logging.Logger) *DatabaseImpl {
	// Badger DB
	options := badger.DefaultOptions(filepath.Join(folder, "db"))
	options.Logger = nil
	db, err := badger.Open(options)
	if err != nil {
		log.Panicf("Error while opening badger database: %s", err.Error())
	}

	// Gorm + SQLite
	sqliteDB, err := gorm.Open(sqlite.Open(filepath.Join(folder, ".db")), &gorm.Config{})
	if err != nil {
		log.Panicf("Error while opening sqlite database: %s", err.Error())
	}
	sqliteDB.AutoMigrate(&storage_models.User{})

	database := &DatabaseImpl{
		db:   db,
		gorm: sqliteDB,
	}

	database.seed()

	return database
}

func (d *DatabaseImpl) GetDB() *badger.DB {
	return d.db
}

func (d *DatabaseImpl) GetGorm() *gorm.DB {
	return d.gorm
}

func (d *DatabaseImpl) seed() {

	admin := storage_models.GetAdminUser()

	var user *storage_models.User
	result := d.GetGorm().First(&user, "email = ?", admin.Email)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		d.GetGorm().Create(admin)
	}
}
