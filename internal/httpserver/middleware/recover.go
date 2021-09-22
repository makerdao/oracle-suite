package middleware

import "net/http"


// Recover recovers from panics that occurred during invoking handler's
// ServeHTTP method.
type Recover struct {
	// Recover is an optional function that will be invoked during panicking.
	// It can be nil.
	Recover func(err interface{})
}

// Handle implements the httpserver.Middleware interface.
func (c *Recover) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil && c.Recover != nil {
				c.Recover(r)
			}
		}()
		next.ServeHTTP(rw, r)
	})
}
