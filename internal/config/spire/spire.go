package spire

import (
	"context"
	"errors"
	"fmt"

	"github.com/makerdao/oracle-suite/pkg/datastore"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/spire"
	"github.com/makerdao/oracle-suite/pkg/transport"
)

var ErrFailedToLoadConfiguration = errors.New("failed to load Spire's configuration")

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
	Feeds     []ethereum.Address
	Logger    log.Logger
}

type ClientDependencies struct {
	Signer ethereum.Signer
}

func (c *Spire) ConfigureAgent(d AgentDependencies) (*spire.Agent, error) {
	agent, err := spire.NewAgent(spire.AgentConfig{
		Datastore: c.configureDatastore(d),
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
	return spire.NewClient(spire.Config{
		Signer:  d.Signer,
		Network: "tcp",
		Address: c.RPC.Address,
	}), nil
}

func (c *Spire) configureDatastore(d AgentDependencies) *datastore.Datastore {
	cfg := datastore.Config{
		Signer:    d.Signer,
		Transport: d.Transport,
		Pairs:     make(map[string]*datastore.Pair),
		Logger:    d.Logger,
	}
	for _, name := range c.Pairs {
		cfg.Pairs[name] = &datastore.Pair{Feeds: d.Feeds}
	}
	return datastore.NewDatastore(cfg)
}
