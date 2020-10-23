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

package oracle

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/makerdao/gofer/internal/ethereum"
)

const priceMultiplier = 1e18

type Price struct {
	AssetPair string
	Val       *big.Int
	Age       time.Time
	V         uint8
	R         [32]byte
	S         [32]byte
}

type jsonPrice struct {
	AssetPair string `json:"ap"`
	Val       string `json:"val"`
	Age       int64  `json:"age"`
	V         uint8  `json:"v"`
	R         string `json:"r"`
	S         string `json:"s"`
}

func NewPrice(assetPair string) *Price {
	return &Price{
		AssetPair: assetPair,
		Val:       nil,
		V:         0,
		R:         [32]byte{},
		S:         [32]byte{},
	}
}

func (p *Price) SetFloat64Price(price float64) {
	pf := new(big.Float).SetFloat64(price)
	pf = new(big.Float).Mul(pf, new(big.Float).SetFloat64(priceMultiplier))
	pi, _ := pf.Int(nil)

	p.Val = pi
}

func (p *Price) Float64Price() float64 {
	x := new(big.Float).SetInt(p.Val)
	x = new(big.Float).Quo(x, new(big.Float).SetFloat64(priceMultiplier))
	f, _ := x.Float64()

	return f
}

func (p *Price) Spread(oldPrice float64) float64 {
	return (p.Float64Price() - oldPrice) / oldPrice * 100
}

func (p *Price) Sign(wallet *ethereum.Wallet) error {
	// Median HEX:
	medianB := make([]byte, 32)
	p.Val.FillBytes(medianB)
	medianHex := hex.EncodeToString(medianB)

	// Time HEX:
	timeHexB := make([]byte, 32)
	binary.BigEndian.PutUint64(timeHexB[24:], uint64(p.Age.Unix()))
	timeHex := hex.EncodeToString(timeHexB)

	// Pair HEX:
	assetPairB := make([]byte, 32)
	copy(assetPairB, p.AssetPair)
	assetPairHex := hex.EncodeToString(assetPairB)

	hash := crypto.Keccak256Hash([]byte("0x" + medianHex + timeHex + assetPairHex))
	sig, err := signData(wallet, hash.Bytes())
	if err != nil {
		return err
	}

	copy(p.R[:], sig[:32])
	copy(p.S[:], sig[32:64])
	p.V = sig[64]

	return nil
}

func (p *Price) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonPrice{
		AssetPair: p.AssetPair,
		Val:       p.Val.String(),
		Age:       p.Age.Unix(),
		V:         p.V,
		R:         hex.EncodeToString(p.R[:]),
		S:         hex.EncodeToString(p.S[:]),
	})
}

func (p *Price) UnmarshalJSON(bytes []byte) error {
	j := &jsonPrice{}
	err := json.Unmarshal(bytes, j)
	if err != nil {
		return err
	}

	p.AssetPair = j.AssetPair
	p.Val, _ = new(big.Int).SetString(j.Val, 10)
	p.Age = time.Unix(j.Age, 0)
	p.V = j.V

	_, err = hex.Decode(p.R[:], []byte(j.R))
	if err != nil {
		return err
	}

	_, err = hex.Decode(p.S[:], []byte(j.S))
	if err != nil {
		return err
	}

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
