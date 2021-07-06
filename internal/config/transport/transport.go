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
