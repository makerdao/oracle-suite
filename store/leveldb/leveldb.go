package leveldb

import (
	"errors"

	"github.com/jmhodges/levigo"
)

type LevelDbStore struct {
	cache     *levigo.Cache
	db        *levigo.DB
	newOpts   *levigo.Options
	writeOpts *levigo.WriteOptions
	readOpts  *levigo.ReadOptions
}

func NewLevelDbStore(cacheSize int) *LevelDbStore {
	cache := levigo.NewLRUCache(cacheSize)

	newOpts := levigo.NewOptions()
	newOpts.SetCache(cache)
	newOpts.SetCreateIfMissing(true)

	return &LevelDbStore{
		cache:     cache,
		newOpts:   newOpts,
		readOpts:  levigo.NewReadOptions(),
		writeOpts: levigo.NewWriteOptions(),
	}
}

func (db *LevelDbStore) IsConnected() bool {
	return db.db != nil
}

func (db *LevelDbStore) GetStatus() error {
	if !db.IsConnected() {
		return errors.New("LevelDB store hasn't been initialized")
	}

	return nil
}

func (db *LevelDbStore) Open(path string) error {
	if db.IsConnected() {
		return errors.New("Can't open, LevelDB store is already open")
	}

	var err error
	db.db, err = levigo.Open(path, db.newOpts)

	return err
}

func (db *LevelDbStore) Close() {
	if !db.IsConnected() {
		return
	}
	db.cache.Close()
	db.db.Close()
	db.db = nil
}

func (db *LevelDbStore) Get(key string) ([]byte, error) {
	if err := db.GetStatus(); err != nil {
		return nil, err
	}
	return db.db.Get(db.readOpts, []byte(key))
}

func (db *LevelDbStore) Put(key string, value []byte) error {
	if err := db.GetStatus(); err != nil {
		return err
	}
	return db.db.Put(db.writeOpts, []byte(key), value)
}

func (db *LevelDbStore) Delete(key string) error {
	if err := db.GetStatus(); err != nil {
		return err
	}
	return db.db.Delete(db.writeOpts, []byte(key))
}
