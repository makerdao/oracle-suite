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

package ethkey

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p-core/crypto"
	cryptoPB "github.com/libp2p/go-libp2p-core/crypto/pb"
	"github.com/libp2p/go-libp2p-core/peer"

	"github.com/makerdao/gofer/pkg/ethereum"
	"github.com/makerdao/gofer/pkg/ethereum/geth"
)

// Eth key type uses the Ethereum keys to sign and verify messages.
const KeyType_Eth cryptoPB.KeyType = 10 //nolint:golint,stylecheck

// NewSigner points to a function which create new Ethereum signer used to
// verify signatures.
var NewSigner = func() ethereum.Signer {
	return geth.NewSigner(nil)
}

func init() {
	crypto.PubKeyUnmarshallers[KeyType_Eth] = UnmarshalEthPublicKey
	crypto.PrivKeyUnmarshallers[KeyType_Eth] = UnmarshalEthPrivateKey
}

// AddressToPeerID converts an Ethereum address to a peer ID. If address is
// invalid then empty ID will be returned.
func AddressToPeerID(a string) peer.ID {
	null := common.Address{}
	addr := common.HexToAddress(a)
	if addr == null {
		return ""
	}
	id, err := peer.IDFromPublicKey(NewPubKey(addr))
	if err != nil {
		return ""
	}
	return id
}

func PeerIDToAddress(id peer.ID) ethereum.Address {
	return common.BytesToAddress([]byte(id))
}
