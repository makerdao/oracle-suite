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
	"errors"
	"net"
	"net/http"
	"net/rpc"

	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/transport"
)

const AgentLoggerTag = "SPIRE_AGENT"

type Agent struct {
	api      *API
	rpc      *rpc.Server
	listener net.Listener
	network  string
	address  string
	log      log.Logger
}

type AgentConfig struct {
	Datastore Datastore
	Transport transport.Transport
	Signer    ethereum.Signer
	Network   string
	Address   string
	Logger    log.Logger
}

func NewAgent(cfg AgentConfig) (*Agent, error) {
	s := &Agent{
		api: &API{
			datastore: cfg.Datastore,
			transport: cfg.Transport,
			signer:    cfg.Signer,
			log:       cfg.Logger.WithField("tag", AgentLoggerTag),
		},
		network: cfg.Network,
		address: cfg.Address,
		log:     cfg.Logger.WithField("tag", AgentLoggerTag),
	}
	s.rpc = rpc.NewServer()
	err := s.rpc.Register(s.api)
	if err != nil {
		return nil, err
	}
	s.rpc.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

	return s, nil
}

func (s *Agent) Start() error {
	s.log.Infof("Starting")
	var err error

	err = s.api.datastore.Start()
	if err != nil {
		return err
	}

	s.log.Debugln("Starting Spire RPC server")

	s.listener, err = net.Listen(s.network, s.address)
	if err != nil {
		return err
	}

	go func() {
		err := http.Serve(s.listener, nil)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.WithError(err).Error("RPC server crashed")
		}
	}()

	return nil
}

func (s *Agent) Stop() {
	defer s.log.Infof("Stopped")
	var err error

	err = s.api.datastore.Stop()
	if err != nil {
		s.log.WithError(err).Error("Unable to stop Datastore")
	}

	err = s.listener.Close()
	if err != nil {
		s.log.WithError(err).Error("Unable to close RPC listener")
	}
}
