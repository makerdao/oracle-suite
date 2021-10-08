//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

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

// HTTPServer allows using middlewares with http.Server and allow controlling
// server lifecycle using context.
type HTTPServer struct {
	ctx    context.Context
	doneCh chan error

	server         *http.Server
	handler        http.Handler
	wrappedHandler http.Handler
	middlewares    []Middleware
}

// New creates a new HTTPServer instance.
func New(ctx context.Context, srv *http.Server) *HTTPServer {
	s := &HTTPServer{
		ctx:    ctx,
		server: srv,
	}
	s.handler = srv.Handler
	srv.Handler = http.HandlerFunc(s.ServeHTTP)
	return s
}

// Use adds a middleware. Middlewares will be called in the order in which they
// were added. This function will panic after calling ServerHTTP/ListenAndServe.
func (s *HTTPServer) Use(m ...Middleware) {
	if s.wrappedHandler != nil {
		panic("cannot add a middleware after calling ServeHTTP/ListenAndServe")
	}
	s.middlewares = append(s.middlewares, m...)
}

// ServeHTTP prepares middlewares stack if necessary and calls ServerHTTP
// on the wrapped server.
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

// ListenAndServe calls function with the same name on the wrapped server.
func (s *HTTPServer) ListenAndServe() error {
	err := s.server.ListenAndServe()
	if err != nil {
		return err
	}
	go s.contextCancelHandler()
	return nil
}

// Wait waits until server is closed.
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
