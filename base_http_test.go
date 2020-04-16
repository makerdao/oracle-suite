package gofer

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/suite"
)

const serverResponse = `{"some":"another","status":1}`
const requiredHeaderKey = "X-Client-Id"
const requiredHeaderValue = "test-client"

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type MakeRequestSuite struct {
	suite.Suite
	server       *httptest.Server
	headerServer *httptest.Server
}

func (suite *MakeRequestSuite) TearDownTest() {
	if suite.server != nil {
		suite.server.Close()
	}
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *MakeRequestSuite) TestMakingRequest() {
	// Start a local HTTP server
	suite.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		rw.Write([]byte(serverResponse))
	}))

	assert.NotNil(suite.T(), suite.server)
	data, err := MakeGetRequest(suite.server.URL)

	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), []byte(serverResponse), data)
}

func (suite *MakeRequestSuite) TestMakingRequestToNotFound() {
	// Start a local HTTP server
	suite.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		rw.WriteHeader(404)
	}))

	assert.NotNil(suite.T(), suite.server)
	data, err := MakeGetRequest(suite.server.URL)

	assert.Error(suite.T(), err)
	assert.Empty(suite.T(), data)

}

func (suite *MakeRequestSuite) TestMakingRequestWithHeaders() {
	// Start a local HTTP server
	suite.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.EqualValues(suite.T(), requiredHeaderValue, req.Header.Get(requiredHeaderKey))
		// Send response to be tested
		rw.Write([]byte(serverResponse))
	}))

	assert.NotNil(suite.T(), suite.server)
	headers := map[string]string{requiredHeaderKey: requiredHeaderValue}
	data, err := MakeGetRequestWithHeaders(suite.server.URL, headers)

	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), []byte(serverResponse), data)
}

func (suite *MakeRequestSuite) TestMakeGetRequestWithRetryFails() {
	calls := 0
	// Start a local HTTP server
	suite.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.EqualValues(suite.T(), requiredHeaderValue, req.Header.Get(requiredHeaderKey))
		calls++
		// Send response to be tested.
		rw.WriteHeader(404)
	}))

	assert.NotNil(suite.T(), suite.server)
	headers := map[string]string{requiredHeaderKey: requiredHeaderValue}
	data, err := MakeGetRequestWithRetry(suite.server.URL, headers, 3)

	assert.Error(suite.T(), err)
	assert.EqualValues(suite.T(), []byte(nil), data)
	assert.EqualValues(suite.T(), 3, calls)
}

func (suite *MakeRequestSuite) TestMakeGetRequestWithRetry() {
	calls := 0
	// Start a local HTTP server
	suite.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.EqualValues(suite.T(), requiredHeaderValue, req.Header.Get(requiredHeaderKey))
		calls++
		// Send response to be tested.
		// Successonly on 3rd call
		if calls < 3 {
			rw.WriteHeader(404)
		} else {
			rw.Write([]byte(serverResponse))
		}
	}))

	assert.NotNil(suite.T(), suite.server)
	headers := map[string]string{requiredHeaderKey: requiredHeaderValue}
	data, err := MakeGetRequestWithRetry(suite.server.URL, headers, 3)

	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), []byte(serverResponse), data)
	assert.EqualValues(suite.T(), 3, calls)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(MakeRequestSuite))
}
