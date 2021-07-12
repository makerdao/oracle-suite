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

	"github.com/sirupsen/logrus"

	"github.com/makerdao/oracle-suite/internal/config"
	goferConfig "github.com/makerdao/oracle-suite/internal/config/gofer"
	"github.com/makerdao/oracle-suite/internal/gofer/marshal"
	pkgGofer "github.com/makerdao/oracle-suite/pkg/gofer"
	"github.com/makerdao/oracle-suite/pkg/gofer/rpc"
	"github.com/makerdao/oracle-suite/pkg/log"
	logLogrus "github.com/makerdao/oracle-suite/pkg/log/logrus"
)

type Config struct {
	Gofer goferConfig.Gofer `json:"gofer"`
}

func (c *Config) Configure(ctx context.Context, logger log.Logger, noRPC bool) (pkgGofer.Gofer, error) {
	return c.Gofer.ConfigureGofer(ctx, logger, noRPC)
}

func (c *Config) ConfigureRPCAgent(ctx context.Context, logger log.Logger) (*rpc.Agent, error) {
	return c.Gofer.ConfigureRPCAgent(ctx, logger)
}

type GoferClientService struct {
	ctxCancel  context.CancelFunc
	Gofer      pkgGofer.Gofer
	Marshaller marshal.Marshaller
}

func PrepareGoferClientService(ctx context.Context, opts *options) (*GoferClientService, error) {
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
	gof, err := opts.Config.Configure(ctx, logger, opts.NoRPC)
	if err != nil {
		return nil, err
	}
	mar, err := marshal.NewMarshal(opts.Format.format)
	if err != nil {
		return nil, err
	}

	return &GoferClientService{
		ctxCancel:  ctxCancel,
		Gofer:      gof,
		Marshaller: mar,
	}, nil
}

func (s *GoferClientService) Start() error {
	if g, ok := s.Gofer.(pkgGofer.StartableGofer); ok {
		return g.Start()
	}
	return nil
}

func (s *GoferClientService) CancelAndWait() {
	s.ctxCancel()
	if g, ok := s.Gofer.(pkgGofer.StartableGofer); ok {
		g.Wait()
	}
}

type GoferAgentService struct {
	ctxCancel context.CancelFunc
	Agent     *rpc.Agent
}

func PrepareGoferAgentService(ctx context.Context, opts *options) (*GoferAgentService, error) {
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
	age, err := opts.Config.ConfigureRPCAgent(ctx, logger)
	if err != nil {
		return nil, err
	}

	return &GoferAgentService{
		ctxCancel: ctxCancel,
		Agent:     age,
	}, nil
}

func (s *GoferAgentService) Start() error {
	return s.Agent.Start()
}

func (s *GoferAgentService) CancelAndWait() {
	s.ctxCancel()
	s.Agent.Wait()
}
