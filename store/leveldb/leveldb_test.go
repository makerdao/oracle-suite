package leveldb

import (
	"os"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type LevelDBStoreTestSuite struct {
    suite.Suite

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
	suite.NoError(openErr, "has opened without error")
	suite.True(suite.db.IsConnected(), "connection open")

	suite.db.Close()
	suite.False(suite.db.IsConnected(), "connection closed")

	_, statErr := os.Stat(suite.path)
	suite.False(os.IsNotExist(statErr), "has created a DB file")
}

func (suite *LevelDBStoreTestSuite) TestPutAndGet() {
	rows := []struct {
		key string
		value []byte
	}{
		{"a", []byte{1} },
		{"b", []byte{2} },
		{"c", []byte{3} },
	}

	// assertions
	openErr := suite.db.Open(suite.path)
	suite.NoError(openErr, "open without error")

	for _, row := range rows {
		putErr := suite.db.Put(row.key, row.value)
		suite.NoError(putErr, "put without error")
	}

	for _, row := range rows {
		buf, getErr := suite.db.Get(row.key)
		suite.NoError(getErr, "get without error")
		suite.Equal(row.value, buf, "Written value same as read value")
	}
}

func (suite *LevelDBStoreTestSuite) TestPutAndDelete() {
	aValue := "value of a"

	// assertions
	openErr := suite.db.Open(suite.path)
	suite.NoError(openErr, "open without error")

	putErr := suite.db.Put("a", []byte(aValue))
	suite.NoError(putErr, "put without error")

	deleteErr := suite.db.Delete("a")
	suite.NoError(deleteErr, "delete without error")

	getReturn, getErr := suite.db.Get("a")
	suite.NoError(getErr, "get without error")
	suite.Nil(getReturn, "get returns nil")
}

func TestLevelDBTestSuite(t *testing.T) {
	suite.Run(t, new(LevelDBStoreTestSuite))
}
