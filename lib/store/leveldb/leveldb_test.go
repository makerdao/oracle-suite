package leveldb

import (
	"os"
	"time"
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"makerdao/gofer/model"
)

type LevelDBStoreTestSuite struct {
    suite.Suite
    suite.SetupAllSuite
    suite.SetupTestSuite
    suite.TearDownTestSuite

    db *LevelDbStore
    path string
    counter int
}

// Before all tests
func (suite *LevelDBStoreTestSuite) SetupSuite() {
	suite.counter = 1
}

// Before each test
func (suite *LevelDBStoreTestSuite) SetupTest() {
	suite.counter += 1
	suite.path = fmt.Sprintf("./test-%d.leveldb", suite.counter)
	suite.db = NewLevelDbStore(3<<30)
}

// After each test
func (suite *LevelDBStoreTestSuite) TearDownTest() {
	suite.db.Close()
	os.RemoveAll(suite.path)
}

func (suite *LevelDBStoreTestSuite) TestOpenAndClose() {
	openErr := suite.db.Open(suite.path)
	suite.Nil(openErr, "has opened without error")

	suite.db.Close()
	_, statErr := os.Stat(suite.path)
	suite.False(os.IsNotExist(statErr), "has created a DB file")
}

func (suite *LevelDBStoreTestSuite) TestPutAndGet() {
	rows := []struct {
		key string
		value *model.Price
	}{
		{"a", &model.Price{ Price: 1 } },
		{"b", &model.Price{ Price: 2 } },
		{"c", &model.Price{ Price: 3 } },
	}

	// assertions
	openErr := suite.db.Open(suite.path)
	suite.Nil(openErr, "open without error")

	for _, row := range rows {
		value, marshalErr := proto.Marshal(row.value)
		suite.Nil(marshalErr, "marshal without error")
		putErr := suite.db.Put(row.key, value)
		suite.Nil(putErr, "put without error")
	}

	for _, row := range rows {
		price := &model.Price{}
		buf, getErr := suite.db.Get(row.key)
		suite.Nil(getErr, "get without error")
		unmarshalErr := proto.Unmarshal(buf, price)
		suite.Nil(unmarshalErr, "unmarshal without error")
		suite.Equal(row.value.Price, price.Price, "Written value same as read value")
	}
}

func (suite *LevelDBStoreTestSuite) TestPutAndDelete() {
	aValue := "value of a"

	// assertions
	openErr := suite.db.Open(suite.path)
	suite.Nil(openErr, "open without error")

	putErr := suite.db.Put("a", []byte(aValue))
	suite.Nil(putErr, "put without error")

	deleteErr := suite.db.Delete("a")
	suite.Nil(deleteErr, "delete without error")

	getReturn, getErr := suite.db.Get("a")
	suite.Nil(getErr, "get without error")
	suite.Nil(getReturn, "get returns nil")

	// tear-down
	suite.db.Close()
	os.RemoveAll(suite.path)
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

func TestLevelDBTestSuite(t *testing.T) {
	suite.Run(t, new(LevelDBStoreTestSuite))
}
