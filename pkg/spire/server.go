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

package spire

import (
	"net"
	"net/http"
	"net/rpc"

	"github.com/makerdao/gofer/pkg/transport"
)

type Server struct {
	api      *API
	rpc      *rpc.Server
	listener net.Listener
	network  string
	address  string
}

func NewServer(datastore Datastore, transport transport.Transport, network, address string) (*Server, error) {
	server := &Server{
		api: &API{
			datastore: datastore,
			transport: transport,
		},
		rpc:     rpc.NewServer(),
		network: network,
		address: address,
	}
	err := server.rpc.Register(server.api)
	if err != nil {
		return nil, err
	}
	server.rpc.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	return server, nil
}

func (s *Server) Start() error {
	var err error
	s.listener, err = net.Listen(s.network, s.address)
	if err != nil {
		return err
	}
	err = s.api.datastore.Start()
	if err != nil {
		return err
	}

	go func() {
		_ = http.Serve(s.listener, nil)
	}()

	return nil
}

func (s *Server) Stop() error {
	err := s.api.datastore.Start()
	if err != nil {
		return err
	}
	return s.listener.Close()
}
