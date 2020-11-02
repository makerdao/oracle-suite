package p2p

import (
	"context"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/multiformats/go-multiaddr"
)

// setupNode creates a new libp2p node with initial peers.
func (p *P2P) setupNode(ctx context.Context, listen string, peers []string) error {
	var err error

	// Start a libp2p node with default settings.
	p.node, err = libp2p.New(ctx, libp2p.ListenAddrStrings(listen))
	if err != nil {
		return err
	}

	// Add initial peers:
	for _, addr := range peers {
		// Turn the destination into a multiaddr.
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return err
		}

		// Extract the peer ID from the multiaddr.
		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return err
		}

		p.node.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	}

	return nil
}
