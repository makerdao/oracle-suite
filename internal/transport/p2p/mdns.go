package p2p

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery"

	"github.com/makerdao/gofer/internal/logger"
)

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "oracles/v3.0.0-alpha"

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	ctx    context.Context
	logger logger.Logger
	host   host.Host
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.logger.Info(LoggerTag, "Discovered new peer %s", pi.String())
	err := n.host.Connect(n.ctx, pi)
	if err != nil {
		n.logger.Info(LoggerTag, "Error connecting to peer %s: %s", pi.String(), err)
	}
}

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func (p *P2P) setupDiscovery() error {
	// Setup mDNS discovery to find local peers.
	disc, err := discovery.NewMdnsService(p.ctx, p.host, DiscoveryInterval, DiscoveryServiceTag)
	if err != nil {
		return err
	}

	n := discoveryNotifee{ctx: p.ctx, logger: p.logger, host: p.host}
	disc.RegisterNotifee(&n)
	return nil
}
