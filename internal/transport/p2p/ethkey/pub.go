package ethkey

import (
	"bytes"
	"errors"

	"github.com/libp2p/go-libp2p-core/crypto"
	crypto_pb "github.com/libp2p/go-libp2p-core/crypto/pb"

	"github.com/makerdao/gofer/internal/ethereum"
)

type PubKey struct {
	address [20]byte
}

func NewPubKey(address [20]byte) crypto.PubKey {
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
func (p *PubKey) Type() crypto_pb.KeyType {
	return KeyType_Eth
}

// Verify implements the crypto.PubKey interface.
func (p *PubKey) Verify(data []byte, sig []byte) (bool, error) {
	// Trim sig to 65 bytes:
	b := make([]byte, 65, 65)
	copy(b, sig)

	// Fetch public address from signature:
	signer := ethereum.NewSigner(nil)
	addr, err := signer.Recover(b, data)
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

	var addr [20]byte
	copy(addr[:], data)
	return &PubKey{address: addr}, nil
}
