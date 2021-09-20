package rpcsplitter

import (
	"net/http"
)

type Handler struct {
	handler http.Handler
}

func NewHandler(endpoints []string) (*Handler, error) {
	rpc, err := newRPC(endpoints)
	if err != nil {
		return nil, err
	}
	return &Handler{handler: &cors{handler: rpc}}, nil
}

func (s *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	s.handler.ServeHTTP(rw, req)
}
