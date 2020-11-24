package node

import (
	"net/rpc"

	"github.com/makerdao/gofer/pkg/transport/messages"
)

type Client struct {
	rpc     *rpc.Client
	network string
	address string
}

func NewClient(network, address string) *Client {
	return &Client{
		network: network,
		address: address,
	}
}

func (s *Client) Start() error {
	client, err := rpc.DialHTTP(s.network, s.address)
	if err != nil {
		return err
	}
	s.rpc = client
	return nil
}

func (s *Client) Stop() error {
	return s.rpc.Close()
}

func (s *Client) BroadcastPrice(price *messages.Price) error {
	err := s.rpc.Call("Api.BroadcastPrice", price, &NoArgument{})
	if err != nil {
		return err
	}
	return nil
}

func (s *Client) GetPrices(assetPair string) ([]*messages.Price, error) {
	prices := &[]*messages.Price{}
	err := s.rpc.Call("Api.GetPrices", assetPair, prices)
	if err != nil {
		return nil, err
	}
	return *prices, nil
}
