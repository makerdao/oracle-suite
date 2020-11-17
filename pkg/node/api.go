package node

import (
	"github.com/makerdao/gofer/internal/transport"
	"github.com/makerdao/gofer/pkg/messages"
)

type NoArgument = struct{}

type Api struct {
	transport transport.Transport
}

func (n *Api) subscribe() error {
	return n.transport.Subscribe(messages.PriceMessageName)
}

func (n *Api) unsubscribe() error {
	return n.transport.Unsubscribe(messages.PriceMessageName)
}

func (n *Api) BroadcastPrice(price *messages.Price, _ *NoArgument) error {
	return n.transport.Broadcast(messages.PriceMessageName, price)
}

func (n *Api) WaitForPrice(_ *NoArgument, price *messages.Price) error {
	return (<-n.transport.WaitFor(messages.PriceMessageName, price)).Error
}
