package rpc

import (
	"net/rpc"

	"github.com/makerdao/gofer/pkg/gofer"
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

func (c *Client) Start() error {
	client, err := rpc.DialHTTP(c.network, c.address)
	if err != nil {
		return err
	}
	c.rpc = client
	return nil
}

func (c *Client) Stop() error {
	return c.rpc.Close()
}

func (c *Client) Nodes(pairs ...gofer.Pair) (map[gofer.Pair]*gofer.Node, error) {
	resp := &NodesResp{}
	err := c.rpc.Call("API.Nodes", NodesArg{Pairs: pairs}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Pairs, nil
}

func (c *Client) Tick(pair gofer.Pair) (*gofer.Tick, error) {
	resp, err := c.Ticks(pair)
	if err != nil {
		return nil, err
	}
	return resp[pair], nil
}

func (c *Client) Ticks(pairs ...gofer.Pair) (map[gofer.Pair]*gofer.Tick, error) {
	resp := &PricesResp{}
	err := c.rpc.Call("API.Prices", PricesArg{Pairs: pairs}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Prices, nil
}

func (c *Client) Pairs() ([]gofer.Pair, error) {
	resp := &PairsResp{}
	err := c.rpc.Call("API.Pairs", &Nothing{}, resp)
	if err != nil {
		return nil, err
	}
	return resp.Pairs, nil
}
