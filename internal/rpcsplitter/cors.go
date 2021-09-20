package rpcsplitter

import (
	"net/http"
)

// cors adds a CORS headers to a handler response.
type cors struct {
	handler http.Handler
}

func (s *cors) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	headers := rw.Header()
	headers.Set("Access-Control-Allow-Origin", req.Header.Get("Origin"))
	switch req.Method {
	case "OPTIONS":
		headers.Set("Access-Control-Allow-Headers", "Content-Type")
		headers.Set("Access-Control-Allow-Methods", "POST")
		headers.Set("Access-Control-Max-Age", "86400")
		rw.WriteHeader(http.StatusNoContent)
	default:
		s.handler.ServeHTTP(rw, req)
	}
}
