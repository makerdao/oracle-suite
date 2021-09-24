package httpserver

import (
	"context"
	"net/http"
	"time"
)

type Middleware interface {
	Handle(http.Handler) http.Handler
}

type MiddlewareFunc func(http.Handler) http.Handler

func (m MiddlewareFunc) Handle(h http.Handler) http.Handler {
	return m(h)
}

type HTTPServer struct {
	ctx    context.Context
	doneCh chan error

	server         *http.Server
	handler        http.Handler
	wrappedHandler http.Handler
	middlewares    []Middleware
}

func New(ctx context.Context, srv *http.Server) *HTTPServer {
	s := &HTTPServer{
		ctx:    ctx,
		server: srv,
	}
	s.handler = srv.Handler
	srv.Handler = http.HandlerFunc(s.ServeHTTP)
	return s
}

func (s *HTTPServer) Use(m ...Middleware) {
	if s.wrappedHandler != nil {
		panic("cannot add a middleware after calling ServerHTTP/ListenAndServe")
	}
	s.middlewares = append(s.middlewares, m...)
}

func (s *HTTPServer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if s.wrappedHandler == nil {
		if len(s.middlewares) == 0 {
			s.wrappedHandler = s.handler
		} else {
			h := s.middlewares[len(s.middlewares)-1].Handle(s.handler)
			for i := len(s.middlewares) - 2; i >= 0; i-- {
				h = s.middlewares[i].Handle(h)
			}
			s.wrappedHandler = h
		}
	}
	s.wrappedHandler.ServeHTTP(rw, r)
}

func (s *HTTPServer) ListenAndServe() error {
	err := s.server.ListenAndServe()
	if err != nil {
		return err
	}
	go s.contextCancelHandler()
	return nil
}

func (s *HTTPServer) Wait() error {
	return <-s.doneCh
}

// contextCancelHandler handles context cancellation.
func (s *HTTPServer) contextCancelHandler() {
	<-s.ctx.Done()
	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second)
	defer ctxCancel()
	s.doneCh <- s.server.Shutdown(ctx)
}
