package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORS_Options(t *testing.T) {
	c := &CORS{
		Origin:  func(r *http.Request) string { return "origin" },
		Headers: func(r *http.Request) string { return "header" },
		Methods: func(r *http.Request) string { return "GET" },
	}
	h := c.Handle(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
	r := httptest.NewRequest("OPTIONS", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusNoContent, rw.Code)
	assert.Equal(t, "origin", rw.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "header", rw.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "GET", rw.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "86400", rw.Header().Get("Access-Control-Allow-Max-Age"))
}

func TestCORS_Get(t *testing.T) {
	c := &CORS{
		Origin:  func(r *http.Request) string { return "origin" },
		Headers: func(r *http.Request) string { return "header" },
		Methods: func(r *http.Request) string { return "GET" },
	}
	h := c.Handle(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "origin", rw.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, rw.Header().Get("Access-Control-Allow-Headers"))
	assert.Empty(t, rw.Header().Get("Access-Control-Allow-Methods"))
	assert.Empty(t, rw.Header().Get("Access-Control-Allow-Max-Age"))
}
