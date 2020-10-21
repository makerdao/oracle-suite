package median

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/makerdao/gofer/pkg/oracle"
)

// TODO: make it configurable
const gasLimit = 200000

type Median struct {
	eth       *oracle.Ethereum
	address   common.Address
	assetPair string
}

func NewMedian(eth *oracle.Ethereum, address common.Address, assetPair string) *Median {
	return &Median{
		eth:       eth,
		address:   address,
		assetPair: assetPair,
	}
}

func (m *Median) Age(ctx context.Context) (time.Time, error) {
	r, err := m.read(ctx, "age")
	if err != nil {
		return time.Unix(0, 0), err
	}

	return time.Unix(r[0].(int64), 0), nil
}

func (m *Median) Bar(ctx context.Context) (int64, error) {
	r, err := m.read(ctx, "bar")
	if err != nil {
		return 0, err
	}

	return r[0].(*big.Int).Int64(), nil
}

func (m *Median) Price(ctx context.Context) (*big.Int, error) {
	b, err := m.eth.Storage(ctx, m.address, common.BigToHash(big.NewInt(1)))
	if err != nil {
		return nil, err
	}
	if len(b) < 32 {
		return nil, errors.New("oracle contract storage query failed")
	}

	return new(big.Int).SetBytes(b[16:32]), err
}

func (m *Median) Poke(ctx context.Context, args []*Price) (*common.Hash, error) {
	// It's important to send prices in correct order, otherwise contract will fail:
	sort.Slice(args, func(i, j int) bool {
		return args[i].Val.Cmp(args[j].Val) < 0
	})

	var (
		val []*big.Int
		age []*big.Int
		v   []uint8
		r   [][32]byte
		s   [][32]byte
	)

	for _, arg := range args {
		if arg.AssetPair != m.assetPair {
			return nil, fmt.Errorf(
				"incompatible asset pair, %s given but %s expected",
				arg.AssetPair,
				m.assetPair,
			)
		}

		val = append(val, arg.Val)
		age = append(age, arg.Age)
		v = append(v, arg.V)
		r = append(r, arg.R)
		s = append(s, arg.S)
	}

	// Simulate transaction to not waste a gas in case of an error:
	if _, err := m.read(ctx, "poke", val, age, v, r, s); err != nil {
		return nil, err
	}

	return m.write(ctx, "poke", val, age, v, r, s)
}

func (m *Median) read(ctx context.Context, method string, args ...interface{}) ([]interface{}, error) {
	cd, err := medianABI.Pack(method, args...)
	if err != nil {
		return nil, err
	}

	data, err := m.eth.Call(ctx, m.address, cd)
	if err != nil {
		return nil, err
	}

	return medianABI.Unpack(method, data)
}

func (m *Median) write(ctx context.Context, method string, args ...interface{}) (*common.Hash, error) {
	cd, err := medianABI.Pack(method, args...)
	if err != nil {
		return nil, err
	}

	return m.eth.SendTransaction(ctx, m.address, gasLimit, cd)
}
