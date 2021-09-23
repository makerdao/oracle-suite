package rpcsplitter

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rpcReq struct {
	ID      int           `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params,omitempty"`
}

type rpcRes struct {
	ID      int         `json:"id"`
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type clientMock struct {
	t *testing.T

	currCall  int
	callMocks []callMock
}

type callMock struct {
	result interface{}
	method string
	params []interface{}
}

func (c *clientMock) mockCall(result interface{}, method string, params ...interface{}) {
	c.callMocks = append(c.callMocks, callMock{
		result: result,
		method: method,
		params: params,
	})
}

func (c *clientMock) Call(result interface{}, method string, params ...interface{}) error {
	if c.currCall >= len(c.callMocks) {
		require.Fail(c.t, "unexpected call")
	}
	defer func() { c.currCall++ }()
	call := c.callMocks[c.currCall]

	assert.Equal(c.t, call.method, method)
	assert.True(c.t, compare(call.params, params))

	if err, ok := call.result.(error); ok {
		return err
	}

	return json.Unmarshal(jsonMarshal(c.t, call.result), &result)
}

type testRPC struct {
	t *testing.T

	clients []rpcClient
	result  interface{}
	method  string
	params  []interface{}
	errors  []string
}

func prepareRPCTest(t *testing.T, clients int, method string, params ...interface{}) *testRPC {
	var cli []rpcClient
	for i := 0; i < clients; i++ {
		cli = append(cli, &clientMock{t: t})
	}
	return &testRPC{t: t, clients: cli, method: method, params: params}
}

func (t *testRPC) mockClientCall(client int, response interface{}, method string, params ...interface{}) *testRPC {
	t.clients[client].(*clientMock).mockCall(response, method, params...)
	return t
}

func (t *testRPC) expectedResult(res interface{}) *testRPC {
	t.result = res
	return t
}

func (t *testRPC) expectedError(msg string) *testRPC {
	t.errors = append(t.errors, msg)
	return t
}

func (t *testRPC) run() {
	// Prepare handler:
	h, err := newHandlerWithClients(t.clients)
	require.NoError(t.t, err)

	// Prepare request:
	id := rand.Int()
	msg := jsonMarshal(t.t, rpcReq{
		ID:      id,
		JSONRPC: "2.0",
		Method:  t.method,
		Params:  t.params,
	})
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(msg))
	r.Header.Set("Content-Type", "application/json")

	// Do request:
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)

	// Parse response:
	res := &rpcRes{}
	jsonUnmarshal(t.t, rw.Body.Bytes(), res)

	// Verify response:
	assert.Equal(t.t, id, res.ID)
	assert.Equal(t.t, "2.0", res.JSONRPC)
	if len(t.errors) > 0 {
		for _, e := range t.errors {
			if e == "" {
				assert.NotEmpty(t.t, res.Error.Message)
			} else {
				assert.Contains(t.t, res.Error.Message, e)
			}
		}
	} else {
		assert.Equal(t.t, 0, res.Error.Code)
		assert.Empty(t.t, res.Error.Message)
		assert.JSONEq(t.t, string(jsonMarshal(t.t, t.result)), string(jsonMarshal(t.t, res.Result)))
	}
}

func jsonMarshal(t *testing.T, v interface{}) []byte {
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func jsonUnmarshal(t *testing.T, b []byte, v interface{}) interface{} {
	require.NoError(t, json.Unmarshal(b, v))
	return v
}
