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

package main

import (
	"context"
	_ "embed"
	"os"

	"github.com/sirupsen/logrus"

	suite "github.com/makerdao/oracle-suite"
	"github.com/makerdao/oracle-suite/internal/config"
	"github.com/makerdao/oracle-suite/pkg/datastore"
	logLogrus "github.com/makerdao/oracle-suite/pkg/log/logrus"
	"github.com/makerdao/oracle-suite/pkg/spire"
	"github.com/makerdao/oracle-suite/pkg/transport"
)

func main() {
	opts := options{Version: suite.Version}
	rootCmd := NewRootCommand(&opts)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type clientServices struct {
	ctxCancel context.CancelFunc
	client    *spire.Client
}

func newClientServices(ctx context.Context, opts *options) (*clientServices, error) {
	var err error
	ctx, ctxCancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			ctxCancel()
		}
	}()

	// Load config file:
	err = config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, err
	}

	// Services:
	cli, err := opts.Config.ConfigureClient(ClientDependencies{
		Context: ctx,
	})
	if err != nil {
		return nil, err
	}

	return &clientServices{
		ctxCancel: ctxCancel,
		client:    cli,
	}, nil
}

func (s *clientServices) start() error {
	var err error
	if err = s.client.Start(); err != nil {
		return err
	}
	return nil
}

func (s *clientServices) cancelAndWait() {
	s.ctxCancel()
	s.client.Wait()
}

type agentServices struct {
	ctxCancel context.CancelFunc
	transport transport.Transport
	datastore datastore.Datastore
	agent     *spire.Agent
}

func newAgentServices(ctx context.Context, opts *options) (*agentServices, error) {
	var err error
	ctx, ctxCancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			ctxCancel()
		}
	}()

	// Load config file:
	err = config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, err
	}

	// Logger:
	ll, err := logrus.ParseLevel(opts.LogVerbosity)
	if err != nil {
		return nil, err
	}
	lr := logrus.New()
	lr.SetLevel(ll)
	lr.SetFormatter(opts.LogFormat.Formatter())
	logger := logLogrus.New(lr)

	// Services:
	tra, dat, age, err := opts.Config.ConfigureAgent(AgentDependencies{
		Context: ctx,
		Logger:  logger,
	})
	if err != nil {
		return nil, err
	}

	return &agentServices{
		ctxCancel: ctxCancel,
		transport: tra,
		datastore: dat,
		agent:     age,
	}, nil
}

func (s *agentServices) start() error {
	var err error
	if err = s.transport.Start(); err != nil {
		return err
	}
	if err = s.datastore.Start(); err != nil {
		return err
	}
	if err = s.agent.Start(); err != nil {
		return err
	}
	return nil
}

func (s *agentServices) cancelAndWait() {
	s.ctxCancel()
	s.transport.Wait()
	s.datastore.Wait()
	s.agent.Wait()
}
