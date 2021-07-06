package main

import (
	"context"

	ethereumConfig "github.com/makerdao/oracle-suite/internal/config/ethereum"
	feedsConfig "github.com/makerdao/oracle-suite/internal/config/feeds"
	spectreConfig "github.com/makerdao/oracle-suite/internal/config/spectre"
	transportConfig "github.com/makerdao/oracle-suite/internal/config/transport"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/spectre"
)

type Config struct {
	P2P      transportConfig.P2P     `json:"p2p"`
	Ethereum ethereumConfig.Ethereum `json:"ethereum"`
	Spectre  spectreConfig.Spectre   `json:"spectre"`
	Feeds    feedsConfig.Feeds       `json:"feeds"`
}

type Dependencies struct {
	Context context.Context
	Logger  log.Logger
}

func (c *Config) Configure(d Dependencies) (*spectre.Spectre, error) {
	sig, err := c.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, err
	}
	cli, err := c.Ethereum.ConfigureEthereumClient(sig)
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
	return c.Spectre.Configure(spectreConfig.Dependencies{
		Signer:         sig,
		Transport:      tra,
		EthereumClient: cli,
		Logger:         d.Logger,
	}), nil
}
