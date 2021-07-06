package transport

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/libp2p/go-libp2p-core/crypto"

	suite "github.com/makerdao/oracle-suite"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/transport"
	"github.com/makerdao/oracle-suite/pkg/transport/messages"
	"github.com/makerdao/oracle-suite/pkg/transport/p2p"
	"github.com/makerdao/oracle-suite/pkg/transport/p2p/crypto/ethkey"
)

var ErrFailedToParsePrivKeySeed = errors.New("failed to parse the privKeySeed field")

type P2P struct {
	PrivKeySeed      string   `json:"privKeySeed"`
	ListenAddrs      []string `json:"listenAddrs"`
	BootstrapAddrs   []string `json:"bootstrapAddrs"`
	DirectPeersAddrs []string `json:"directPeersAddrs"`
	BlockedAddrs     []string `json:"blockedAddrs"`
	DisableDiscovery bool     `json:"disableDiscovery"`
}

type Dependencies struct {
	Context context.Context
	Signer  ethereum.Signer
	Feeds   []ethereum.Address
	Logger  log.Logger
}

func (c *P2P) Configure(d Dependencies) (transport.Transport, error) {
	peerPrivKey, err := c.generatePrivKey()
	if err != nil {
		return nil, err
	}
	cfg := p2p.Config{
		Context:          d.Context,
		PeerPrivKey:      peerPrivKey,
		MessagePrivKey:   ethkey.NewPrivKey(d.Signer),
		ListenAddrs:      c.ListenAddrs,
		BootstrapAddrs:   c.BootstrapAddrs,
		DirectPeersAddrs: c.DirectPeersAddrs,
		BlockedAddrs:     c.BlockedAddrs,
		FeedersAddrs:     d.Feeds,
		Discovery:        !c.DisableDiscovery,
		Signer:           d.Signer,
		Logger:           d.Logger,
		AppName:          "spire",
		AppVersion:       suite.Version,
	}
	p, err := p2p.New(cfg)
	if err != nil {
		return nil, err
	}
	err = p.Subscribe(messages.PriceMessageName, (*messages.Price)(nil))
	if err != nil {
		_ = p.Close()
		return nil, err
	}
	return p, nil
}

func (c *P2P) generatePrivKey() (crypto.PrivKey, error) {
	seedReader := rand.Reader
	if len(c.PrivKeySeed) != 0 {
		seed, err := hex.DecodeString(c.PrivKeySeed)
		if err != nil {
			return nil, fmt.Errorf("%v: %v", ErrFailedToParsePrivKeySeed, err)
		}
		if len(seed) != ed25519.SeedSize {
			return nil, fmt.Errorf("%v: seed must be 32 bytes", ErrFailedToParsePrivKeySeed)
		}
		seedReader = bytes.NewReader(seed)
	}
	privKey, _, err := crypto.GenerateEd25519Key(seedReader)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrFailedToParsePrivKeySeed, err)
	}
	return privKey, nil
}
