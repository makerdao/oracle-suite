package rpc

import (
	"net/rpc"

	"github.com/makerdao/gofer/pkg/gofer"
)

// RPC implements the gofer.Gofer interface. It uses a remote RPC server to
// fetch prices and models.
type RPC struct {
	rpc     *rpc.Client
	network string
	address string
}

// NewRPC returns a new RPC instance.
func NewRPC(network, address string) *RPC {
	return &RPC{
		network: network,
		address: address,
	}
}

// Start implements the gofer.StartableGofer interface.
func (c *RPC) Start() error {
	client, err := rpc.DialHTTP(c.network, c.address)
	if err != nil {
		return err
	}
	c.rpc = client
	return nil
}

// Stop implements the gofer.StartableGofer interface.
func (c *RPC) Stop() error {
	return c.rpc.Close()
}

// Models implements the gofer.Gofer interface.
func (c *RPC) Models(pairs ...gofer.Pair) (map[gofer.Pair]*gofer.Model, error) {
	resp := &NodesResp{}
	err := c.rpc.Call("API.Models", NodesArg{Pairs: pairs}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Pairs, nil
}

// Price implements the gofer.Gofer interface.
func (c *RPC) Price(pair gofer.Pair) (*gofer.Price, error) {
	resp, err := c.Prices(pair)
	if err != nil {
		return nil, err
	}
	return resp[pair], nil
}

// Prices implements the gofer.Gofer interface.
func (c *RPC) Prices(pairs ...gofer.Pair) (map[gofer.Pair]*gofer.Price, error) {
	resp := &PricesResp{}
	err := c.rpc.Call("API.Prices", PricesArg{Pairs: pairs}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Prices, nil
}

// Pairs implements the gofer.Gofer interface.
func (c *RPC) Pairs() ([]gofer.Pair, error) {
	resp := &PairsResp{}
	err := c.rpc.Call("API.Pairs", &Nothing{}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Pairs, nil
}
