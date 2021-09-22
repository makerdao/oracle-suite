package middleware

import (
	"net/http"
	"time"

	"github.com/makerdao/oracle-suite/pkg/log"
)

const httpRequestLog = "HTTP request"

// Logger prints logs for each request. If the log level is set to debug, it
// will print the contents of requests and responses.
type Logger struct {
	// Log is an instance of a log.Logger. It cannot be nil, otherwise code will panic.
	Log log.Logger
}

// Handle implements the httpserver.Middleware interface.
func (l *Logger) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		t := time.Now()
		e := l.Log
		if l.Log.Level() >= log.Debug {
			rw = newRecorder(rw)
			e = e.WithField("request", string(readRequest(r)))
		}
		defer func() {
			e = e.WithFields(log.Fields{
				"duration": time.Since(t),
				"method":   r.Method,
				"url":      r.URL.String(),
			})
			if l.Log.Level() >= log.Debug {
				e = e.WithFields(log.Fields{
					"response": string(readResponse(rw.(*recorder))),
					"status":   rw.(*recorder).code,
				})
				e.Debug(httpRequestLog)
			} else {
				e.Info(httpRequestLog)
			}
		}()
		next.ServeHTTP(rw, r)
	})
}
