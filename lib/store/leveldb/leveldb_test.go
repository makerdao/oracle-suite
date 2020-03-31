package leveldb

import (
	"os"
	"time"
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"makerdao/gofer/model"

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
		value *model.Price
	}{
		{"a", &model.Price{ Price: 1 } },
		{"b", &model.Price{ Price: 2 } },
		{"c", &model.Price{ Price: 3 } },
	}
	db := NewLevelDbStore(3<<30)

	// assertions
	openErr := db.Open(path)
	assert.Nil(t, openErr, "open without error")

	for _, row := range rows {
		value, marshalErr := proto.Marshal(row.value)
		assert.Nil(t, marshalErr, "marshal without error")
		putErr := db.Put(row.key, value)
		assert.Nil(t, putErr, "put without error")
	}

	for _, row := range rows {
		price := &model.Price{}
		buf, getErr := db.Get(row.key)
		assert.Nil(t, getErr, "get without error")
		unmarshalErr := proto.Unmarshal(buf, price)
		assert.Nil(t, unmarshalErr, "unmarshal without error")
		assert.Equal(t, row.value.Price, price.Price, "Written value same as read value")
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

func BenchmarkListRW(b *testing.B) {
	// setup
	path := "./benchmark-list.leveldb"
	db := NewLevelDbStore(3<<30)
	_ = db.Open(path)

	// assertions

	for i := 0; i < b.N; i++ {
		price := &model.Price{
			Timestamp: time.Now().Unix(),
			Exchange: "mjaucoinmarket",
			Base: "bench",
			Quote: "mark",
			Price: 1337 + float64(i),
			Volume: 0.1 + float64(i),
		}
		buf, _ := proto.Marshal(price)
		_ = db.Put(fmt.Sprintf("list-%d", i), buf)
	}

	for i := 0; i < b.N; i++ {
		price := &model.Price{}
		buf, _ := db.Get(fmt.Sprintf("list-%f", math.Floor(rand.Float64() * float64(b.N))))
		_ = proto.Unmarshal(buf, price)
	}

	// tear-down
	db.Close()
	os.RemoveAll(path)
}
