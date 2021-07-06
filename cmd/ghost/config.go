package main

import (
	"context"

	ethereumConfig "github.com/makerdao/oracle-suite/internal/config/ethereum"
	feedsConfig "github.com/makerdao/oracle-suite/internal/config/feeds"
	ghostConfig "github.com/makerdao/oracle-suite/internal/config/ghost"
	goferConfig "github.com/makerdao/oracle-suite/internal/config/gofer"
	transportConfig "github.com/makerdao/oracle-suite/internal/config/transport"
	"github.com/makerdao/oracle-suite/pkg/ghost"

	"github.com/makerdao/oracle-suite/pkg/log"
)

type Config struct {
	Gofer    goferConfig.Gofer       `json:"gofer"`
	Ethereum ethereumConfig.Ethereum `json:"ethereum"`
	P2P      transportConfig.P2P     `json:"p2p"`
	Ghost    ghostConfig.Ghost       `json:"ghost"`
	Feeds    feedsConfig.Feeds       `json:"feeds"`
}

type Dependencies struct {
	Context context.Context
	Logger  log.Logger
}

func (c *Config) Configure(d Dependencies) (*ghost.Ghost, error) {
	gof, err := c.Gofer.ConfigureGofer(d.Logger, true)
	if err != nil {
		return nil, err
	}
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
	return c.Ghost.Configure(ghostConfig.Dependencies{
		Context:   d.Context,
		Gofer:     gof,
		Signer:    sig,
		Transport: tra,
		Logger:    d.Logger,
	})
}
