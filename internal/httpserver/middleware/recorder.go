package middleware

import (
	"bytes"
	"io"
	"net/http"
)

// recorder implements the http.ResponseWriter interface. It passes all calls
// to the underlying ResponseWriter and records a copies of values for a later
// inspection.
type recorder struct {
	rw      http.ResponseWriter // rw is an underlying ResponseWriter.
	code    int                 // code is the HTTP status code
	headers http.Header         // headers is the list of HTTP headers
	body    *bytes.Buffer       // body is the HTTP response body
}

func newRecorder(rw http.ResponseWriter) *recorder {
	return &recorder{
		rw:      rw,
		headers: make(http.Header),
		body:    new(bytes.Buffer),
		code:    http.StatusOK,
	}
}

func (r *recorder) Header() http.Header {
	return r.rw.Header()
}

func (r *recorder) Write(buf []byte) (int, error) {
	r.body.Write(buf)
	return r.rw.Write(buf)
}

func (r *recorder) WriteHeader(code int) {
	r.code = code
	r.rw.WriteHeader(code)
}

func readRequest(r *http.Request) []byte {
	b, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(b))
	return b
}

func readResponse(r *recorder) []byte {
	b, _ := io.ReadAll(r.body)
	return b
}
