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

package rpc

import (
	"errors"
	"net"
	"net/http"
	"net/rpc"

	"github.com/makerdao/gofer/pkg/gofer"
	"github.com/makerdao/gofer/pkg/gofer/local"
	"github.com/makerdao/gofer/pkg/gofer/local/feeder"
	"github.com/makerdao/gofer/pkg/gofer/local/graph"
	"github.com/makerdao/gofer/pkg/log"
)

const ServerLoggerTag = "GOFER_SERVER"

type Server struct {
	api      *API
	rpc      *rpc.Server
	graphs   map[gofer.Pair]graph.Aggregator
	feeder   *feeder.Feeder
	listener net.Listener
	network  string
	address  string
	log      log.Logger
}

type ServerConfig struct {
	Graphs  map[gofer.Pair]graph.Aggregator
	Feeder  *feeder.Feeder
	Network string
	Address string
	Logger  log.Logger
}

func NewServer(cfg ServerConfig) (*Server, error) {
	server := &Server{
		api: &API{
			gofer: local.NewGofer(cfg.Graphs, nil),
			log:   cfg.Logger.WithField("tag", ServerLoggerTag),
		},
		rpc:     rpc.NewServer(),
		graphs:  cfg.Graphs,
		feeder:  cfg.Feeder,
		network: cfg.Network,
		address: cfg.Address,
		log:     cfg.Logger.WithField("tag", ServerLoggerTag),
	}
	err := server.rpc.Register(server.api)
	if err != nil {
		return nil, err
	}
	server.rpc.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	return server, nil
}

func (s *Server) Start() error {
	s.log.Infof("Starting")

	var err error
	s.listener, err = net.Listen(s.network, s.address)
	if err != nil {
		return err
	}

	var nodes []graph.Node
	for _, node := range s.graphs {
		nodes = append(nodes, node)
	}
	err = s.feeder.Start(nodes...)
	if err != nil {
		return err
	}

	go func() {
		err := http.Serve(s.listener, nil)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Error(err.Error())
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	defer s.log.Infof("Stopped")
	s.feeder.Stop()
	return s.listener.Close()
}
