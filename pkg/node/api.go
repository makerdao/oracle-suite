package node

import (
	"github.com/makerdao/gofer/pkg/datastore"
	"github.com/makerdao/gofer/pkg/transport"
	"github.com/makerdao/gofer/pkg/transport/messages"
)

type NoArgument = struct{}

type Datastore interface {
	Prices() *datastore.PriceStore
	Start() error
	Stop() error
}

type Api struct {
	transport transport.Transport
	datastore Datastore
}

func (n *Api) BroadcastPrice(price *messages.Price, _ *NoArgument) error {
	return n.transport.Broadcast(messages.PriceMessageName, price)
}

func (n *Api) GetPrices(assetPair *string, prices *[]*messages.Price) error {
	*prices = n.datastore.Prices().AssetPair(*assetPair).Messages()
	return nil
}
