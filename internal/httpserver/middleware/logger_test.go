package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/log/callback"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger_InfoLevel(t *testing.T) {
	var recordedLogs []log.Fields
	m := callback.New(log.Info, func(level log.Level, fields log.Fields, msg string) {
		if level != log.Info {
			return
		}
		fields["_msg"] = msg
		delete(fields, "duration")
		recordedLogs = append(recordedLogs, fields)
	})
	h := (&Logger{Log: m}).Handle(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)

	require.Len(t, recordedLogs, 1)
	assert.Equal(t, log.Fields{
		"_msg":   httpRequestLog,
		"method": "GET",
		"url":    "/",
	}, recordedLogs[0])
}

func TestLogger_DebugLevel(t *testing.T) {
	var recordedLogs []log.Fields
	m := callback.New(log.Debug, func(level log.Level, fields log.Fields, msg string) {
		if level != log.Debug {
			return
		}
		fields["_msg"] = msg
		delete(fields, "duration")
		recordedLogs = append(recordedLogs, fields)
	})
	h := (&Logger{Log: m}).Handle(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("response"))
		writer.WriteHeader(http.StatusNotFound)
	}))
	r := httptest.NewRequest("GET", "/", strings.NewReader("request"))
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)

	require.Len(t, recordedLogs, 1)
	assert.Equal(t, log.Fields{
		"_msg":     httpRequestLog,
		"method":   "GET",
		"url":      "/",
		"status":   http.StatusNotFound,
		"request":  "request",
		"response": "response",
	}, recordedLogs[0])
}
