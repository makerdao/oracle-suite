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
	"bytes"
	"errors"

	"github.com/libp2p/go-libp2p-core/crypto"
	cryptoPB "github.com/libp2p/go-libp2p-core/crypto/pb"

	"github.com/makerdao/gofer/pkg/ethereum"
)

type PubKey struct {
	address ethereum.Address
}

func NewPubKey(address ethereum.Address) crypto.PubKey {
	return &PubKey{
		address: address,
	}
}

// Bytes implements the crypto.Key interface.
func (p *PubKey) Bytes() ([]byte, error) {
	return crypto.MarshalPublicKey(p)
}

// Equals implements the crypto.Key interface.
func (p *PubKey) Equals(key crypto.Key) bool {
	if p.Type() != key.Type() {
		return false
	}

	a, err := p.Raw()
	if err != nil {
		return false
	}
	b, err := key.Raw()
	if err != nil {
		return false
	}

	return bytes.Equal(a, b)
}

// Raw implements the crypto.Key interface.
func (p *PubKey) Raw() ([]byte, error) {
	return p.address[:], nil
}

// Type implements the crypto.Key interface.
func (p *PubKey) Type() cryptoPB.KeyType {
	return KeyType_Eth
}

// Verify implements the crypto.PubKey interface.
func (p *PubKey) Verify(data []byte, sig []byte) (bool, error) {
	// Trim sig to 65 bytes:
	b := make([]byte, 65, 65)
	copy(b, sig)

	// Fetch public address from signature:
	addr, err := NewSigner(nil).Recover(b, data)
	if err != nil {
		return false, err
	}

	// Verify address:
	return bytes.Equal(addr.Bytes(), p.address[:]), nil
}

// UnmarshalEthPublicKey returns a public key from input bytes.
func UnmarshalEthPublicKey(data []byte) (crypto.PubKey, error) {
	if len(data) != 20 {
		return nil, errors.New("expect eth public key data size to be 20")
	}

	var addr ethereum.Address
	copy(addr[:], data)
	return &PubKey{address: addr}, nil
}
