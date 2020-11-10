package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/multiformats/go-multiaddr"

	"github.com/makerdao/gofer/internal/log"
	"github.com/makerdao/gofer/internal/transport"
	"github.com/makerdao/gofer/internal/transport/p2p/banner"
	"github.com/makerdao/gofer/internal/transport/p2p/logger"
)

const LoggerTag = "P2P"

type P2P struct {
	log    log.Logger
	node   *Node
	banner *banner.Banner
}

type Config struct {
	Context        context.Context
	Logger         log.Logger
	PrivKey        string
	ListenAddrs    []string
	BootstrapPeers []string
	BannedPeers    []string
}

func NewP2P(config Config) (*P2P, error) {
	var err error

	l := log.WrapLogger(config.Logger, log.Fields{"tag": LoggerTag})
	p := &P2P{
		log:  l,
		node: NewNode(config.Context, l),
	}

	logger.Register(p.node, l)
	p.banner = banner.Register(p.node)

	err = p.startNode(config.PrivKey, config.ListenAddrs)
	if err != nil {
		return nil, err
	}
	err = p.bannedPeers(config.BannedPeers)
	if err != nil {
		return nil, err
	}
	err = p.bootstrapPeers(config.BootstrapPeers)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Subscribe implements the transport.Transport interface.
func (p *P2P) Subscribe(topic string) error {
	return p.node.Subscribe(topic)
}

// Unsubscribe implements the transport.Transport interface.
func (p *P2P) Unsubscribe(topic string) error {
	return p.node.Unsubscribe(topic)
}

// Broadcast implements the transport.Transport interface.
func (p *P2P) Broadcast(topic string, message transport.Message) error {
	sub, err := p.node.Subscription(topic)
	if err != nil {
		return err
	}
	return sub.publish(message)
}

// WaitFor implements the transport.Transport interface.
func (p *P2P) WaitFor(topic string, message transport.Message) chan transport.Status {
	sub, err := p.node.Subscription(topic)
	if err != nil {
		return nil
	}
	return sub.next(message)
}

// Close implements the transport.Transport interface.
func (p *P2P) Close() error {
	defer p.log.Info("Stopped")
	return p.node.Close()
}

func (p *P2P) startNode(pkStr string, addrs []string) error {
	if len(addrs) == 0 {
		addrs = []string{"/ip4/0.0.0.0/tcp/0"}
	}

	// Parse multiaddresses strings.
	var maddrs []multiaddr.Multiaddr
	for _, addrstr := range addrs {
		maddr, err := multiaddr.NewMultiaddr(addrstr)
		if err != nil {
			return err
		}
		maddrs = append(maddrs, maddr)
	}

	// Convert PK from string. If string is empty, ignore and pass nil
	// as a value.
	var ppk *crypto.PrivKey
	if pkStr != "" {
		pkBts, err := crypto.ConfigDecodeKey(pkStr)
		if err != nil {
			return err
		}
		pk, err := crypto.UnmarshalPrivateKey(pkBts)
		if err != nil {
			return err
		}
		ppk = &pk
	}

	// Connect to P2P network.
	err := p.node.Start(ppk, maddrs)
	if err != nil {
		return err
	}

	// Print log with node's parameters.
	p.log.
		WithFields(log.Fields{"addrs": p.listenAddrStrs()}).
		Info("Listening")

	return nil
}

func (p *P2P) bannedPeers(addrs []string) error {
	for _, addrstr := range addrs {
		maddr, err := multiaddr.NewMultiaddr(addrstr)
		if err != nil {
			return err
		}

		err = p.banner.Ban(maddr)
		if err != nil {
			p.log.
				WithFields(log.Fields{"addr": addrstr}).
				WithError(err).
				Warn("Unable to ban given address")
		}
	}
	return nil
}

func (p *P2P) bootstrapPeers(addrs []string) error {
	for _, addrstr := range addrs {
		maddr, err := multiaddr.NewMultiaddr(addrstr)
		if err != nil {
			return err
		}

		err = p.node.Connect(maddr)
		if err != nil {
			p.log.
				WithFields(log.Fields{"addr": addrstr}).
				WithError(err).
				Warn("Unable to connect to bootstrap peer")
		}
	}
	return nil
}

func (p *P2P) listenAddrStrs() []string {
	var addrstrs []string
	for _, addr := range p.node.Addrs() {
		addrstrs = append(addrstrs, fmt.Sprintf("%s/p2p/%s", addr.String(), p.node.ID()))
	}
	return addrstrs
}
