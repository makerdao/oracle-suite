package datastore

import (
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/transport/messages"
)

type FeederPrice struct {
	AssetPair string
	Feeder    ethereum.Address
}

type PriceStore interface {
	// Add adds a new price to the list. If a price from same feeder already
	// exists, the newer one will be used.
	Add(from ethereum.Address, msg *messages.Price)
	// All returns all prices.
	All() map[FeederPrice]*messages.Price
	// AssetPair returns all prices for given asset pair.
	AssetPair(assetPair string) []*messages.Price
	// Feeder returns the latest price for given asset pair sent by given feeder.
	Feeder(assetPair string, feeder ethereum.Address) *messages.Price
}
