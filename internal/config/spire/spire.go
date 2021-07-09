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
	"context"
	"errors"
	"fmt"

	"github.com/makerdao/oracle-suite/pkg/datastore"
	datastoreMemory "github.com/makerdao/oracle-suite/pkg/datastore/memory"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/spire"
	"github.com/makerdao/oracle-suite/pkg/transport"
)

var ErrFailedToLoadConfiguration = errors.New("failed to load Spire's configuration")

//nolint:unlambda
var spireAgentFactory = func(cfg spire.AgentConfig) (*spire.Agent, error) {
	return spire.NewAgent(cfg)
}

//nolint:unlambda
var spireClientFactory = func(cfg spire.ClientConfig) (*spire.Client, error) {
	return spire.NewClient(cfg)
}

//nolint:unlambda
var datastoreFactory = func(cfg datastoreMemory.Config) (datastore.Datastore, error) {
	return datastoreMemory.NewDatastore(cfg)
}

type Spire struct {
	RPC   RPC      `json:"rpc"`
	Pairs []string `json:"pairs"`
}

type RPC struct {
	Address string `json:"address"`
}

type AgentDependencies struct {
	Context   context.Context
	Signer    ethereum.Signer
	Transport transport.Transport
	Datastore datastore.Datastore
	Feeds     []ethereum.Address
	Logger    log.Logger
}

type ClientDependencies struct {
	Context context.Context
	Signer  ethereum.Signer
}

type DatastoreDependencies struct {
	Context   context.Context
	Signer    ethereum.Signer
	Transport transport.Transport
	Feeds     []ethereum.Address
	Logger    log.Logger
}

func (c *Spire) ConfigureAgent(d AgentDependencies) (*spire.Agent, error) {
	agent, err := spireAgentFactory(spire.AgentConfig{
		Context:   d.Context,
		Datastore: d.Datastore,
		Transport: d.Transport,
		Signer:    d.Signer,
		Network:   "tcp",
		Address:   c.RPC.Address,
		Logger:    d.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrFailedToLoadConfiguration, err)
	}
	return agent, nil
}

func (c *Spire) ConfigureClient(d ClientDependencies) (*spire.Client, error) {
	return spireClientFactory(spire.ClientConfig{
		Context: d.Context,
		Signer:  d.Signer,
		Network: "tcp",
		Address: c.RPC.Address,
	})
}

func (c *Spire) ConfigureDatastore(d DatastoreDependencies) (datastore.Datastore, error) {
	cfg := datastoreMemory.Config{
		Context:   d.Context,
		Signer:    d.Signer,
		Transport: d.Transport,
		Pairs:     make(map[string]*datastoreMemory.Pair),
		Logger:    d.Logger,
	}
	for _, name := range c.Pairs {
		cfg.Pairs[name] = &datastoreMemory.Pair{Feeds: d.Feeds}
	}
	return datastoreFactory(cfg)
}
