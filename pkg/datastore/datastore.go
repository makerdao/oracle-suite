package datastore

import (
	"errors"
	"math/big"

	"github.com/makerdao/gofer/internal/ethereum"
	"github.com/makerdao/gofer/internal/log"
	"github.com/makerdao/gofer/internal/transport"
	"github.com/makerdao/gofer/pkg/messages"
)

const LoggerTag = "DATASTORE"

// Datastore reads and stores prices from the P2P network.
type Datastore struct {
	signer     ethereum.Signer
	transport  transport.Transport
	pairs      map[string]*Pair
	priceStore *PriceStore
	log        log.Logger
	doneCh     chan struct{}
}

type Config struct {
	// Signer is an instance of the ethereum.Signer which will be used to
	// verify price signatures.
	Signer ethereum.Signer
	// Transport is a implementation of transport used to fetch prices from
	// feeders.
	Transport transport.Transport
	// Pairs is the list supported pairs by Spectre with their configuration.
	Pairs map[string]*Pair
	// Interval describes how often we should try to update Oracles.
	// Logger is a current logger interface used by the Datastore. The Logger is
	// required to monitor asynchronous processes.
	Logger log.Logger
}

type Pair struct {
	// Feeds is the list of Ethereum addresses from which prices will be
	// accepted.
	Feeds []ethereum.Address
}

func NewDatastore(config Config) *Datastore {
	return &Datastore{
		signer:     config.Signer,
		transport:  config.Transport,
		pairs:      config.Pairs,
		priceStore: NewPriceStore(),
		log:        log.WrapLogger(config.Logger, log.Fields{"tag": LoggerTag}),
		doneCh:     make(chan struct{}),
	}
}

func (c *Datastore) Start() error {
	c.log.Info("Starting")

	err := c.transport.Subscribe(messages.PriceMessageName)
	if err != nil {
		return err
	}

	return c.collectorLoop()
}

func (c *Datastore) Stop() error {
	defer c.log.Info("Stopped")

	close(c.doneCh)
	err := c.transport.Unsubscribe(messages.PriceMessageName)
	if err != nil {
		return err
	}

	return nil
}

func (c *Datastore) Prices() *PriceStore {
	return c.priceStore
}

// collectPrice adds a price from a feeder which may be used to update
// Oracle contract. The price will be added only if a feeder is
// allowed to send prices (must be on the r.Feeds list).
func (c *Datastore) collectPrice(msg *messages.Price) error {
	from, err := msg.Price.From(c.signer)
	if err != nil {
		return errors.New("received price has an invalid signature")
	}
	if _, ok := c.pairs[msg.Price.AssetPair]; !ok {
		return errors.New("received pair is not configured")
	}
	if !c.isFeedAllowed(msg.Price.AssetPair, *from) {
		return errors.New("feeder is not allowed to send prices")
	}
	if msg.Price.Val.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("received price is invalid")
	}

	c.priceStore.Add(*from, msg)

	return nil
}

// collectorLoop creates a asynchronous loop which fetches prices from feeders.
func (c *Datastore) collectorLoop() error {
	go func() {
		for {
			price := &messages.Price{}
			select {
			case <-c.doneCh:
				return
			case status := <-c.transport.WaitFor(messages.PriceMessageName, price):
				// If there was a problem while reading prices from the transport:
				if status.Error != nil {
					c.log.
						WithError(status.Error).
						Warn("Unable to read prices from the transport")
					continue
				}

				// Try to collect received price:
				err := c.collectPrice(price)

				// Prepare log fields:
				from, _ := price.Price.From(c.signer)
				fields := log.Fields{"assetPair": price.Price.AssetPair}
				if from != nil {
					fields["from"] = from.String()
				}

				// Print logs:
				if err != nil {
					c.log.
						WithError(err).
						WithFields(fields).
						Warn("Received invalid price")
				} else {
					c.log.
						WithFields(fields).
						Info("Price received")
				}
			}
		}
	}()

	return nil
}

func (c *Datastore) isFeedAllowed(assetPair string, address ethereum.Address) bool {
	for _, a := range c.pairs[assetPair].Feeds {
		if a == address {
			return true
		}
	}
	return false
}