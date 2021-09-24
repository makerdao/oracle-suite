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
	l := callback.New(log.Info, func(level log.Level, fields log.Fields, msg string) {
		if level != log.Info {
			return
		}
		delete(fields, "duration")
		fields["_msg"] = msg
		recordedLogs = append(recordedLogs, fields)
	})

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h := (&Logger{Log: l}).Handle(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
	h.ServeHTTP(w, r)

	require.Len(t, recordedLogs, 1)
	assert.Equal(t, log.Fields{
		"_msg":   httpRequestLog,
		"method": "GET",
		"url":    "/",
	}, recordedLogs[0])
}

func TestLogger_DebugLevel(t *testing.T) {
	var recordedLogs []log.Fields
	l := callback.New(log.Debug, func(level log.Level, fields log.Fields, msg string) {
		if level != log.Debug {
			return
		}
		delete(fields, "duration")
		fields["_msg"] = msg
		recordedLogs = append(recordedLogs, fields)
	})

	r := httptest.NewRequest("GET", "/", strings.NewReader("request"))
	w := httptest.NewRecorder()
	h := (&Logger{Log: l}).Handle(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("response"))
		writer.WriteHeader(http.StatusNotFound)
	}))
	h.ServeHTTP(w, r)

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
