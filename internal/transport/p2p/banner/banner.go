package banner

import (
	"net"

	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
)

type Node interface {
	AddConnectionGater(connGaters ...connmgr.ConnectionGater)
	PubSub() *pubsub.PubSub
}

// Register registers extensions to P2P node which will allow to ban nodes
// using theirs IPs or IDs.
func Register(node Node) *Banner {
	banner := &Banner{
		connGater: &connGater{},
		pubSub:    node.PubSub(),
	}

	node.AddConnectionGater(banner.connGater)

	return banner
}

type Banner struct {
	connGater *connGater
	pubSub    *pubsub.PubSub
}

func (b *Banner) Ban(maddr multiaddr.Multiaddr) error {
	multiaddr.ForEach(maddr, func(c multiaddr.Component) bool {
		switch c.Protocol().Code {
		case multiaddr.P_IP4:
			b.BanIP(net.ParseIP(c.String()))
		case multiaddr.P_IP6:
			b.BanIP(net.ParseIP(c.String()))
		case multiaddr.P_P2P:
			id, err := peer.IDFromBytes(c.RawValue())
			if err != nil {
				return true
			}
			b.BanPeer(id)
		}
		return true
	})

	return nil
}

func (b *Banner) BanIP(ip net.IP) {
	b.connGater.BanIP(ip)
}

func (b *Banner) BanPeer(id peer.ID) {
	b.pubSub.BlacklistPeer(id)
}
