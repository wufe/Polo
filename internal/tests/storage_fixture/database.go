package storage_fixture

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/wufe/polo/pkg/storage"
)

type FixtureDBOptions struct {
	Clean bool
}

func NewDB(folder string, options *FixtureDBOptions) *storage.Database {

	if options == nil {
		options = &FixtureDBOptions{}
	}

	database := storage.NewDB(folder)

	if options.Clean {
		cleanup(database)
	}

	return database
}

func cleanup(database *storage.Database) {
	err := database.DB.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := txn.Delete(item.Key())
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}
