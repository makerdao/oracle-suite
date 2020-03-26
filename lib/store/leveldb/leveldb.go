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
	closed    bool
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
		closed:    false,
	}
}

func (db *LevelDbStore) IsOpen() (bool, error) {
	if db.closed {
		return false, errors.New("LevelDB store is closed")
	}

	if db.db == nil {
		return false, errors.New("LevelDB store hasn't been initialized")
	}

	return true, nil
}

func (db *LevelDbStore) Open(path string) error {
	if db.closed {
		return errors.New("LevelDB store is closed")
	}

	db_, err := levigo.Open(path, db.newOpts)
	db.db = db_

	if err != nil {
		db.closed = true
	}

	return err
}

func (db *LevelDbStore) Close() error {
	if ok, err := db.IsOpen(); !ok {
		return err
	}
	db.cache.Close()
	db.db.Close()
	db.closed = true
	return nil
}

func (db *LevelDbStore) Get(key string) ([]byte, error) {
	if ok, err := db.IsOpen(); !ok {
		return nil, err
	}
	return db.db.Get(db.readOpts, []byte(key))
}

func (db *LevelDbStore) Put(key string, value []byte) error {
	if ok, err := db.IsOpen(); !ok {
		return err
	}
	return db.db.Put(db.writeOpts, []byte(key), value)
}

func (db *LevelDbStore) Delete(key string) error {
	if ok, err := db.IsOpen(); !ok {
		return err
	}
	return db.db.Delete(db.writeOpts, []byte(key))
}
