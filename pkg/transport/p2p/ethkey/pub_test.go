package ethkey

import (
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/pkg/ethereum"
	"github.com/makerdao/gofer/pkg/ethereum/mocks"
)

func TestPubKey_Equals(t *testing.T) {
	pub1 := NewPubKey(testAddress1)
	pub2 := NewPubKey(testAddress2)

	assert.True(t, pub1.Equals(pub1))
	assert.False(t, pub1.Equals(pub2))
}

func TestPubKey_Raw(t *testing.T) {
	pub := NewPubKey(testAddress1)

	b, err := pub.Raw()
	assert.NoError(t, err)
	assert.Equal(t, testAddress1.Bytes(), b)
}

func TestPubKey_Type(t *testing.T) {
	assert.Equal(t, KeyType_Eth, NewPubKey(testAddress1).Type())
}

func TestPubKey_Verify(t *testing.T) {
	sig := &mocks.Signer{}
	ethSig := ethereum.SignatureFromBytes([]byte("bar"))
	orgSigner := NewSigner
	NewSigner = func() ethereum.Signer { return sig }
	pub := NewPubKey(testAddress1)
	bts := []byte("foo")

	// Valid:
	sig.On("Recover", ethSig, []byte("foo")).Return(&testAddress1, nil).Once()
	ok, err := pub.Verify(bts, ethSig.Bytes())
	assert.True(t, ok)
	assert.NoError(t, err)

	// Invalid:
	sig.On("Recover", ethSig, []byte("foo")).Return(&testAddress2, nil).Once()
	ok, err = pub.Verify(bts, ethSig.Bytes())
	assert.False(t, ok)
	assert.NoError(t, err)

	// Error:
	sig.On("Recover", ethSig, []byte("foo")).Return((*common.Address)(nil), errors.New("err")).Once()
	ok, err = pub.Verify(bts, ethSig.Bytes())
	assert.False(t, ok)
	assert.Error(t, err)

	NewSigner = orgSigner
}
