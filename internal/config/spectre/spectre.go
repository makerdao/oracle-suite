package spectre

import (
	"time"

	"github.com/makerdao/oracle-suite/pkg/datastore"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/log"
	oracleGeth "github.com/makerdao/oracle-suite/pkg/oracle/geth"
	"github.com/makerdao/oracle-suite/pkg/spectre"
	"github.com/makerdao/oracle-suite/pkg/transport"
)

type Spectre struct {
	Interval    int                   `json:"interval"`
	Medianizers map[string]Medianizer `json:"medianizers"`
}

type Medianizer struct {
	Contract         string  `json:"oracle"`
	OracleSpread     float64 `json:"oracleSpread"`
	OracleExpiration int64   `json:"oracleExpiration"`
	MsgExpiration    int64   `json:"msgExpiration"`
}

type Dependencies struct {
	Signer         ethereum.Signer
	Transport      transport.Transport
	EthereumClient ethereum.Client
	Feeds          []ethereum.Address
	Logger         log.Logger
}

func (c *Spectre) Configure(d Dependencies) *spectre.Spectre {
	cfg := spectre.Config{
		Signer:    d.Signer,
		Interval:  time.Second * time.Duration(c.Interval),
		Datastore: c.configureDatastore(d),
		Logger:    d.Logger,
	}
	for name, pair := range c.Medianizers {
		cfg.Pairs = append(cfg.Pairs, &spectre.Pair{
			AssetPair:        name,
			OracleSpread:     pair.OracleSpread,
			OracleExpiration: time.Second * time.Duration(pair.OracleExpiration),
			PriceExpiration:  time.Second * time.Duration(pair.MsgExpiration),
			Median:           oracleGeth.NewMedian(d.EthereumClient, ethereum.HexToAddress(pair.Contract)),
		})
	}
	return spectre.NewSpectre(cfg)
}

func (c *Spectre) configureDatastore(d Dependencies) *datastore.Datastore {
	cfg := datastore.Config{
		Signer:    d.Signer,
		Transport: d.Transport,
		Pairs:     make(map[string]*datastore.Pair),
		Logger:    d.Logger,
	}
	for name := range c.Medianizers {
		cfg.Pairs[name] = &datastore.Pair{Feeds: d.Feeds}
	}
	return datastore.NewDatastore(cfg)
}
