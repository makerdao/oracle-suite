package main

import (
	"context"

	ethereumConfig "github.com/makerdao/oracle-suite/internal/config/ethereum"
	feedsConfig "github.com/makerdao/oracle-suite/internal/config/feeds"
	spireConfig "github.com/makerdao/oracle-suite/internal/config/spire"
	transportConfig "github.com/makerdao/oracle-suite/internal/config/transport"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/spire"
)

type Config struct {
	P2P      transportConfig.P2P     `json:"p2p"`
	Ethereum ethereumConfig.Ethereum `json:"ethereum"`
	Spire    spireConfig.Spire       `json:"spire"`
	Feeds    feedsConfig.Feeds       `json:"feeds"`
}

type Dependencies struct {
	Context context.Context
	Logger  log.Logger
}

func (c *Config) ConfigureClient() (*spire.Client, error) {
	sig, err := c.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, err
	}
	return c.Spire.ConfigureClient(spireConfig.ClientDependencies{
		Signer: sig,
	})
}

func (c *Config) ConfigureAgent(d Dependencies) (*spire.Agent, error) {
	sig, err := c.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, err
	}
	tra, err := c.P2P.Configure(transportConfig.Dependencies{
		Context: d.Context,
		Signer:  sig,
		Feeds:   c.Feeds.Addresses(),
		Logger:  d.Logger,
	})
	if err != nil {
		return nil, err
	}
	return c.Spire.ConfigureAgent(spireConfig.AgentDependencies{
		Context:   d.Context,
		Signer:    sig,
		Transport: tra,
		Feeds:     c.Feeds.Addresses(),
		Logger:    d.Logger,
	})
}
