package median

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/makerdao/gofer/pkg/oracle"
)

// TODO: make it configurable
const gasLimit = 200000

type Median struct {
	eth       *oracle.Ethereum
	abi       abi.ABI
	address   common.Address
	assetPair string
}

type PokeVal struct {
	Val *big.Int
	Age *big.Int
	V   uint8
	R   [32]byte
	S   [32]byte
}

func (m *PokeVal) Sign(wallet *oracle.Wallet, assetPair string) error {
	// Median HEX:
	medianB := make([]byte, 32)
	m.Val.FillBytes(medianB)
	medianHex := hex.EncodeToString(medianB)

	// Time HEX:
	timeHexB := make([]byte, 32)
	binary.BigEndian.PutUint64(timeHexB[24:], uint64(time.Now().Unix()))
	timeHex := hex.EncodeToString(timeHexB)

	// Pair HEX:
	assetPairB := make([]byte, 32)
	copy(assetPairB, assetPair)
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

func NewMedian(eth *oracle.Ethereum, address common.Address, assetPair string) (*Median, error) {
	medianABI, err := abi.JSON(strings.NewReader(medianABI))
	if err != nil {
		return nil, err
	}

	return &Median{
		eth:       eth,
		abi:       medianABI,
		address:   address,
		assetPair: assetPair,
	}, nil
}

func (m *Median) Age(ctx context.Context) (uint32, error) {
	r, err := m.read(ctx, "age")
	if err != nil {
		return 0, err
	}

	return r[0].(uint32), nil
}

func (m *Median) Bar(ctx context.Context) (*big.Int, error) {
	r, err := m.read(ctx, "bar")
	if err != nil {
		return nil, err
	}

	return r[0].(*big.Int), nil
}

func (m *Median) Price(ctx context.Context) (*big.Int, error) {
	b, err := m.eth.Storage(ctx, m.address, common.BigToHash(big.NewInt(1)))
	if len(b) < 48 {
		return nil, errors.New("oracle contract storage query failed")
	}

	return new(big.Int).SetBytes(b[16:32]), err
}

func (m *Median) Poke(ctx context.Context, wallet *oracle.Wallet, args []PokeVal) (*common.Hash, error) {
	if len(args) == 0 {
		return nil, errors.New("poke requires at least one value")
	}

	var (
		val []*big.Int
		age []*big.Int
		v   []uint8
		r   [][32]byte
		s   [][32]byte
	)

	for _, arg := range args {
		val = append(val, arg.Val)
		age = append(age, arg.Age)
		v = append(v, arg.V)
		r = append(r, arg.R)
		s = append(s, arg.S)
	}

	return m.write(ctx, "poke", val, age, v, r, s)
}

func (m *Median) read(ctx context.Context, method string) ([]interface{}, error) {
	cd, err := m.abi.Pack(method)
	if err != nil {
		return nil, err
	}

	data, err := m.eth.Call(ctx, m.address, cd)
	if err != nil {
		return nil, err
	}

	return m.abi.Unpack(method, data)
}

func (m *Median) write(ctx context.Context, method string, args ...interface{}) (*common.Hash, error) {
	cd, err := m.abi.Pack(method, args...)
	if err != nil {
		return nil, err
	}

	return m.eth.SendTransaction(ctx, m.address, gasLimit, cd)
}


func signData(w *oracle.Wallet, data []byte) ([]byte, error) {
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
