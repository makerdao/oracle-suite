package leveldb

import (
	"os"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenAndClose(t *testing.T) {
	// setup
	path := "./test-open-close.leveldb"
	db := NewLevelDbStore(3<<30)

	// assertions
	openErr := db.Open(path)
	assert.Nil(t, openErr, "has opened without error")

	db.Close()
	_, statErr := os.Stat(path)
	assert.False(t, os.IsNotExist(statErr), "has created a DB file")

	// tear-down
	os.RemoveAll(path)
}

func TestPutAndGet(t *testing.T) {
	// setup
	path := "./test-put-get.leveldb"
	rows := []struct {
		key string
		value string
	}{
		{"a", "value of a"},
		{"b", "value of b"},
		{"c", "value of c"},
	}
	db := NewLevelDbStore(3<<30)

	// assertions
	openErr := db.Open(path)
	assert.Nil(t, openErr, "open without error")

	for _, row := range rows {
		putErr := db.Put(row.key, []byte(row.value))
		assert.Nil(t, putErr, "put without error")
	}

	for _, row := range rows {
		getValue, getErr := db.Get(row.key)
		assert.Nil(t, getErr, "get without error")
		assert.Equal(t, row.value, string(getValue), "Written value same as read value")
	}

	// tear-down
	db.Close()
	os.RemoveAll(path)
}

func TestPutAndDelete(t *testing.T) {
	// setup
	path := "./test-put-delete.leveldb"
	aValue := "value of a"

	db := NewLevelDbStore(3<<30)

	// assertions
	openErr := db.Open(path)
	assert.Nil(t, openErr, "open without error")

	putErr := db.Put("a", []byte(aValue))
	assert.Nil(t, putErr, "put without error")

	deleteErr := db.Delete("a")
	assert.Nil(t, deleteErr, "delete without error")

	getReturn, getErr := db.Get("a")
	assert.Nil(t, getErr, "get without error")
	assert.Nil(t, getReturn, "get returns nil")

	// tear-down
	db.Close()
	os.RemoveAll(path)
}
