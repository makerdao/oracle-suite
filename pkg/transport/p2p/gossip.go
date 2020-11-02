package p2p

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// setupGossip creates a new PubSub service using the GossipSub router.
func (p *P2P) setupGossip(ctx context.Context) error {
	var err error

	p.ps, err = pubsub.NewGossipSub(ctx, p.node)
	if err != nil {
		return err
	}

	return nil
}
