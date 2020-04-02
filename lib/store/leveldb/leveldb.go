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

func (db *LevelDbStore) GetStatus() string {
	if db.closed {
		return "LevelDB store is closed"
	}

	if db.db == nil {
		return "LevelDB store hasn't been initialized"
	}

	return "LevelDB store is open"
}

func (db *LevelDbStore) IsOpen() bool {
	if db.closed {
		return false
	}

	if db.db == nil {
		return false
	}

	return true
}

func (db *LevelDbStore) Open(path string) error {
	var err error

	if db.closed {
		return errors.New("Can't open, LevelDB store has been closed")
	}

	db.db, err = levigo.Open(path, db.newOpts)
	if err != nil {
		db.closed = true
	}

	return err
}

func (db *LevelDbStore) Close() {
	if db.IsOpen() {
		db.cache.Close()
		db.db.Close()
		db.closed = true
	}
}

func (db *LevelDbStore) Get(key string) ([]byte, error) {
	if !db.IsOpen() {
		return nil, errors.New(db.GetStatus())
	}
	return db.db.Get(db.readOpts, []byte(key))
}

func (db *LevelDbStore) Put(key string, value []byte) error {
	if !db.IsOpen() {
		return errors.New(db.GetStatus())
	}
	return db.db.Put(db.writeOpts, []byte(key), value)
}

func (db *LevelDbStore) Delete(key string) error {
	if !db.IsOpen() {
		return errors.New(db.GetStatus())
	}
	return db.db.Delete(db.writeOpts, []byte(key))
}
