package middleware

import "net/http"

type Recover struct {
	Recover func(err interface{})
}

func (c *Recover) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				if c.Recover != nil {
					c.Recover(r)
				}
			}
		}()
		next.ServeHTTP(rw, r)
	})
}
