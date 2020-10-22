package oracle

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/makerdao/gofer/internal/ethereum"
)

type Price struct {
	AssetPair string
	Val       *big.Int
	Age       time.Time
	V         uint8
	R         [32]byte
	S         [32]byte
}

func (m *Price) Sign(wallet *ethereum.Wallet) error {
	// Median HEX:
	medianB := make([]byte, 32)
	m.Val.FillBytes(medianB)
	medianHex := hex.EncodeToString(medianB)

	// Time HEX:
	timeHexB := make([]byte, 32)
	binary.BigEndian.PutUint64(timeHexB[24:], uint64(m.Age.Unix()))
	timeHex := hex.EncodeToString(timeHexB)

	// Pair HEX:
	assetPairB := make([]byte, 32)
	copy(assetPairB, m.AssetPair)
	assetPairHex := hex.EncodeToString(assetPairB)

	hash := crypto.Keccak256Hash([]byte("0x" + medianHex + timeHex + assetPairHex))
	sig, err := signData(wallet, hash.Bytes())
	if err != nil {
		return err
	}

	copy(m.R[:], sig[:32])
	copy(m.S[:], sig[32:64])
	m.V = sig[64]

	return nil
}

func signData(w *ethereum.Wallet, data []byte) ([]byte, error) {
	msg := []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data))
	wallet := w.EthWallet()
	account := w.EthAccount()

	signature, err := wallet.SignDataWithPassphrase(*account, w.Passphrase(), "", msg)
	if err != nil {
		return nil, err
	}

	// Transform V from 0/1 to 27/28 according to the yellow paper.
	signature[64] += 27

	return signature, nil
}
