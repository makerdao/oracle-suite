package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer_WithoutMiddlewares(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	srv := New(context.Background(), &http.Server{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("response"))
		}),
	})

	srv.ServeHTTP(rw, r)
	assert.Equal(t, "response", rw.Body.String())
}

func TestServer_WithMiddlewares(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	srv := New(context.Background(), &http.Server{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("response"))
		}),
	})
	srv.Use(MiddlewareFunc(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("before-"))
			handler.ServeHTTP(rw, r)
		})
	}))
	srv.Use(MiddlewareFunc(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(rw, r)
			rw.Write([]byte("-after"))
		})
	}))

	srv.ServeHTTP(rw, r)
	assert.Equal(t, "before-response-after", rw.Body.String())
}

func TestServer_UseMustPanicAfterServerHTTP(t *testing.T) {
	var panicVal interface{}
	defer func() { panicVal = recover() }()

	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	srv := New(context.Background(), &http.Server{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}),
	})
	srv.ServeHTTP(rw, r)
	srv.Use(MiddlewareFunc(func(handler http.Handler) http.Handler { return nil }))

	assert.NotNil(t, panicVal)
}
