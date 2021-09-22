package middleware

import (
	"net/http"
)

// CORS adds a basic support for CORS preflight requests.
type CORS struct {
	// Origin is a function that returns a value of
	// an Access-Control-Allow-Origin header. It can not be nil.
	Origin func(r *http.Request) string
	// Headers is a function that returns a value of
	// an Access-Control-Allow-Headers header. It can not be nil.
	Headers func(r *http.Request) string
	// Methods is a function that returns a value of
	// an Access-Control-Allow-Methods. It can not be nil.
	Methods func(r *http.Request) string
}

// Handle implements the httpserver.Middleware interface.
func (c *CORS) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		headers := rw.Header()
		headers.Set("Access-Control-Allow-Origin", c.Origin(r))
		switch r.Method {
		case "OPTIONS":
			headers.Set("Access-Control-Allow-Headers", c.Headers(r))
			headers.Set("Access-Control-Allow-Methods", c.Methods(r))
			headers.Set("Access-Control-Max-Age", "86400")
			rw.WriteHeader(http.StatusNoContent)
		default:
			next.ServeHTTP(rw, r)
		}
	})
}
