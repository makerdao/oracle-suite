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
	ctx            context.Context
	server         *http.Server
	handler        http.Handler
	wrappedHandler http.Handler
	middlewares    []Middleware
	doneCh         chan error
}

func New(ctx context.Context, srv *http.Server) *HTTPServer {
	s := &HTTPServer{
		ctx:    ctx,
		server: srv,
	}
	s.handler = srv.Handler
	s.wrappedHandler = srv.Handler
	srv.Handler = http.HandlerFunc(s.serveHTTP)
	return s
}

func (s *HTTPServer) Use(m ...Middleware) {
	s.middlewares = append(s.middlewares, m...)
	s.wrappedHandler = s.wrapHandler(s.handler)
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

func (s *HTTPServer) serveHTTP(rw http.ResponseWriter, r *http.Request) {
	s.wrappedHandler.ServeHTTP(rw, r)
}

func (s *HTTPServer) wrapHandler(final http.Handler) http.Handler {
	if len(s.middlewares) == 0 {
		return final
	}
	h := s.middlewares[len(s.middlewares)-1].Handle(final)
	for i := len(s.middlewares) - 2; i >= 0; i-- {
		h = s.middlewares[i].Handle(h)
	}
	return h
}
