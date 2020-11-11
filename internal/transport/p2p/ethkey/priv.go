package ethkey

import (
	"bytes"
	"errors"

	"github.com/libp2p/go-libp2p-core/crypto"
	crypto_pb "github.com/libp2p/go-libp2p-core/crypto/pb"

	"github.com/makerdao/gofer/internal/ethereum"
)

type PrivKey struct {
	wallet *ethereum.Wallet
}

func NewPrivKey(wallet *ethereum.Wallet) crypto.PrivKey {
	return &PrivKey{
		wallet: wallet,
	}
}

// Bytes implements the crypto.Key interface.
func (p *PrivKey) Bytes() ([]byte, error) {
	return crypto.MarshalPrivateKey(p)
}

// Equals implements the crypto.Key interface.
func (p *PrivKey) Equals(key crypto.Key) bool {
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
func (p *PrivKey) Raw() ([]byte, error) {
	return p.wallet.Address().Bytes(), nil
}

// Type implements the crypto.Key interface.
func (p *PrivKey) Type() crypto_pb.KeyType {
	return KeyType_Eth
}

// Sign implements the crypto.PrivKey interface.
func (p *PrivKey) Sign(bytes []byte) ([]byte, error) {
	return ethereum.NewSigner(p.wallet).Signature(bytes)
}

// GetPublic implements the crypto.PrivKey interface.
func (p *PrivKey) GetPublic() crypto.PubKey {
	return NewPubKey(p.wallet.Address())
}

// UnmarshalEthPrivateKey should return private key from input bytes, but this
// not supported for ethereum keys.
func UnmarshalEthPrivateKey(data []byte) (crypto.PrivKey, error) {
	return nil, errors.New("eth key type does not support unmarshalling")
}
