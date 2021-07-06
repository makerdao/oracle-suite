package ghost

import (
	"context"
	"time"

	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/ghost"
	"github.com/makerdao/oracle-suite/pkg/gofer"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/transport"
)

type Ghost struct {
	Interval int      `json:"interval"`
	Pairs    []string `json:"pairs"`
}

type Dependencies struct {
	Context   context.Context
	Gofer     gofer.Gofer
	Signer    ethereum.Signer
	Transport transport.Transport
	Logger    log.Logger
}

func (c *Ghost) Configure(d Dependencies) (*ghost.Ghost, error) {
	cfg := ghost.Config{
		Gofer:     d.Gofer,
		Signer:    d.Signer,
		Transport: d.Transport,
		Logger:    d.Logger,
		Interval:  time.Second * time.Duration(c.Interval),
		Pairs:     nil,
	}
	for _, name := range c.Pairs {
		cfg.Pairs = append(cfg.Pairs, &ghost.Pair{
			AssetPair: name,
		})
	}
	return ghost.NewGhost(cfg)
}
