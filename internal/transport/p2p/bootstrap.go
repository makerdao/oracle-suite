package p2p

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

func (p *P2P) bootstrapNodes(peers []string) error {
	// Add bootstrap nodes:
	for _, addr := range peers {
		// Turn the destination into a multiaddr.
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return err
		}

		// Extract the peer ID from the multiaddr.
		pi, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return err
		}

		p.logger.Info(LoggerTag, "Bootstrap peer %s", pi.String())
		err = p.host.Connect(p.ctx, *pi)
		if err != nil {
			p.logger.Info(LoggerTag, "Error connecting to peer %s: %s", pi.String(), err)
		}
	}

	return nil
}
