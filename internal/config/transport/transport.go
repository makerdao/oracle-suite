//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package transport

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/libp2p/go-libp2p-core/crypto"

	suite "github.com/makerdao/oracle-suite"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/transport"
	"github.com/makerdao/oracle-suite/pkg/transport/libp2p"
	"github.com/makerdao/oracle-suite/pkg/transport/libp2p/crypto/ethkey"
	"github.com/makerdao/oracle-suite/pkg/transport/messages"
)

var libP2PTransportFactory = func(ctx context.Context, cfg libp2p.Config) (transport.Transport, error) {
	return libp2p.New(ctx, cfg)
}

type Transport struct {
	LibP2P LibP2P      `json:"libp2p"`
	SSB    Scuttlebutt `json:"ssb"`
}

type LibP2P struct {
	PrivKeySeed      string   `json:"privKeySeed"`
	ListenAddrs      []string `json:"listenAddrs"`
	BootstrapAddrs   []string `json:"bootstrapAddrs"`
	DirectPeersAddrs []string `json:"directPeersAddrs"`
	BlockedAddrs     []string `json:"blockedAddrs"`
	DisableDiscovery bool     `json:"disableDiscovery"`
}

type Scuttlebutt struct {
	Caps string `json:"caps"`
}

type Caps struct {
	Shs    string `json:"shs"`
	Sign   string `json:"sign"`
	Invite string `json:"invite,omitempty"`
}

type LibP2PDependencies struct {
	Context context.Context
	Signer  ethereum.Signer
	Feeds   []ethereum.Address
	Logger  log.Logger
}

type BootstrapDependencies struct {
	Context context.Context
	Logger  log.Logger
}

func (c *Transport) ConfigureSSB() (transport.Transport, error) {
	return nil, nil
}
func (c *Transport) ConfigureLibP2P(d LibP2PDependencies) (transport.Transport, error) {
	peerPrivKey, err := c.generatePrivKey()
	if err != nil {
		return nil, err
	}
	cfg := libp2p.Config{
		Mode:             libp2p.ClientMode,
		PeerPrivKey:      peerPrivKey,
		Topics:           map[string]transport.Message{messages.PriceMessageName: (*messages.Price)(nil)},
		MessagePrivKey:   ethkey.NewPrivKey(d.Signer),
		ListenAddrs:      c.LibP2P.ListenAddrs,
		BootstrapAddrs:   c.LibP2P.BootstrapAddrs,
		DirectPeersAddrs: c.LibP2P.DirectPeersAddrs,
		BlockedAddrs:     c.LibP2P.BlockedAddrs,
		FeedersAddrs:     d.Feeds,
		Discovery:        !c.LibP2P.DisableDiscovery,
		Signer:           d.Signer,
		Logger:           d.Logger,
		AppName:          "spire",
		AppVersion:       suite.Version,
	}
	p, err := libP2PTransportFactory(d.Context, cfg)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (c *Transport) ConfigureP2PBoostrap(d BootstrapDependencies) (transport.Transport, error) {
	peerPrivKey, err := c.generatePrivKey()
	if err != nil {
		return nil, err
	}
	cfg := libp2p.Config{
		Mode:             libp2p.BootstrapMode,
		PeerPrivKey:      peerPrivKey,
		ListenAddrs:      c.LibP2P.ListenAddrs,
		BootstrapAddrs:   c.LibP2P.BootstrapAddrs,
		DirectPeersAddrs: c.LibP2P.DirectPeersAddrs,
		BlockedAddrs:     c.LibP2P.BlockedAddrs,
		Logger:           d.Logger,
		AppName:          "bootstrap",
		AppVersion:       suite.Version,
	}
	p, err := libP2PTransportFactory(d.Context, cfg)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (c *Transport) generatePrivKey() (crypto.PrivKey, error) {
	seedReader := rand.Reader
	if len(c.LibP2P.PrivKeySeed) != 0 {
		seed, err := hex.DecodeString(c.LibP2P.PrivKeySeed)
		if err != nil {
			return nil, fmt.Errorf("invalid privKeySeed value, failed to decode hex data: %w", err)
		}
		if len(seed) != ed25519.SeedSize {
			return nil, fmt.Errorf("invalid privKeySeed value, 32 bytes expected")
		}
		seedReader = bytes.NewReader(seed)
	}
	privKey, _, err := crypto.GenerateEd25519Key(seedReader)
	if err != nil {
		return nil, fmt.Errorf("invalid privKeySeed value, failed to generate key: %w", err)
	}
	return privKey, nil
}
