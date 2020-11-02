package p2p

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// setupGossip creates a new PubSub service using the GossipSub router.
func (p *P2P) setupGossip() error {
	var err error

	p.ps, err = pubsub.NewGossipSub(p.ctx, p.node)
	if err != nil {
		return err
	}

	return nil
}
