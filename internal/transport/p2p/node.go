package p2p

import (
	"context"

	"github.com/libp2p/go-libp2p"
)

// setupNode creates a new libp2p node with initial peers.
func (p *P2P) setupNode(ctx context.Context, listen string) error {
	var err error

	// Start a libp2p node with default settings.
	p.host, err = libp2p.New(ctx, libp2p.ListenAddrStrings(listen))
	if err != nil {
		return err
	}

	return nil
}
