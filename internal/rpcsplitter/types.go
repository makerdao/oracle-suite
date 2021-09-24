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

package rpcsplitter

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
)

const addressLength = 20
const hashLength = 32

// jsonType stores an argument in its raw form and passes it to
// endpoints unchanged.
type jsonType struct{ j interface{} }

func newJSON(j string) *jsonType {
	t := &jsonType{}
	err := t.UnmarshalJSON([]byte(j))
	if err != nil {
		return nil
	}
	return t
}

// MarshalJSON returns m as the JSON encoding of m.
func (t jsonType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.j)
}

// UnmarshalJSON sets *m to a copy of data.
func (t *jsonType) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &t.j)
}

func (t *jsonType) Compare(v interface{}) bool {
	if v, ok := v.(*jsonType); ok {
		return compare(t.j, v.j)
	}
	return false
}

type blockNumberType big.Int

const earliestBlockNumber = -1
const latestBlockNumber = -2
const pendingBlockNumber = -3

func newBlockNumber(n string) *blockNumberType {
	b := &blockNumberType{}
	if err := b.UnmarshalJSON([]byte(n)); err != nil {
		return nil
	}
	return b
}

// MarshalJSON implements json.Marshaler.
func (n blockNumberType) MarshalJSON() ([]byte, error) {
	switch {
	case n.IsEarliest():
		return []byte(`"earliest"`), nil
	case n.IsLatest():
		return []byte(`"latest"`), nil
	case n.IsPending():
		return []byte(`"pending"`), nil
	default:
		return naiveQuote(bigIntToHex((*big.Int)(&n))), nil
	}
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *blockNumberType) UnmarshalJSON(input []byte) error {
	input = naiveUnquote(input)
	switch strings.TrimSpace(string(input)) {
	case "earliest":
		*n = *((*blockNumberType)(big.NewInt(earliestBlockNumber)))
		return nil
	case "latest":
		*n = *((*blockNumberType)(big.NewInt(latestBlockNumber)))
		return nil
	case "pending":
		*n = *((*blockNumberType)(big.NewInt(pendingBlockNumber)))
		return nil
	default:
		u, err := hexToBigInt(input)
		if err != nil {
			return err
		}
		if u.Cmp(big.NewInt(math.MaxInt64)) > 0 {
			return fmt.Errorf("block number larger than int64")
		}
		*n = blockNumberType(*u)
		return nil
	}
}

func (n *blockNumberType) Compare(v interface{}) bool {
	if v, ok := v.(*blockNumberType); ok {
		return n.Big().Cmp(v.Big()) == 0
	}
	return false
}

func (n *blockNumberType) IsEarliest() bool {
	return n.Big().Int64() == earliestBlockNumber
}

func (n *blockNumberType) IsLatest() bool {
	return n.Big().Int64() == latestBlockNumber
}

func (n *blockNumberType) IsPending() bool {
	return n.Big().Int64() == pendingBlockNumber
}

func (n *blockNumberType) IsTag() bool {
	return n.Big().Sign() < 0
}

func (n *blockNumberType) Big() *big.Int {
	return (*big.Int)(n)
}

type numberType big.Int

func newNumber(n string) *numberType {
	b, ok := new(big.Int).SetString(strings.TrimPrefix(n, "0x"), 16)
	if !ok {
		return nil
	}
	return (*numberType)(b)
}

// MarshalJSON implements json.Marshaler.
func (n numberType) MarshalJSON() ([]byte, error) {
	return naiveQuote(bigIntToHex((*big.Int)(&n))), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *numberType) UnmarshalJSON(input []byte) error {
	u, err := hexToBigInt(naiveUnquote(input))
	if err != nil {
		return err
	}
	*n = numberType(*u)
	return nil
}

func (n *numberType) Compare(v interface{}) bool {
	if v, ok := v.(*numberType); ok {
		return n.Big().Cmp(v.Big()) == 0
	}
	return false
}

func (n *numberType) Big() *big.Int {
	return (*big.Int)(n)
}

// bytesType marshals/unmarshals as a JSON string with 0x prefix.
// The empty slice marshals as "0x".
type bytesType []byte

func newBytes(hex string) bytesType {
	b, err := hexToBytes([]byte(hex))
	if err != nil {
		return nil
	}
	return b
}

// MarshalJSON implements json.Marshaler.
func (b bytesType) MarshalJSON() ([]byte, error) {
	return naiveQuote(bytesToHex(b)), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *bytesType) UnmarshalJSON(input []byte) error {
	if bytes.Equal(input, []byte("null")) {
		return nil
	}
	u, err := hexToBytes(naiveUnquote(input))
	*b = u
	return err
}

func (b *bytesType) Compare(v interface{}) bool {
	if v, ok := v.(*bytesType); ok {
		return bytes.Equal(*b, *v)
	}
	return false
}

// addressType marshals/unmarshals as an Ethereum address.
type addressType [addressLength]byte

func newAddress(address string) addressType {
	a := addressType{}
	b, err := hexToBytes([]byte(address))
	if err != nil {
		return a
	}
	if len(b) != addressLength {
		return a
	}
	copy(a[:], b)
	return a
}

// MarshalJSON implements json.Marshaler.
func (b addressType) MarshalJSON() ([]byte, error) {
	return naiveQuote(bytesToHex(b[:])), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *addressType) UnmarshalJSON(input []byte) error {
	u, err := hexToBytes(naiveUnquote(input))
	if len(u) != addressLength {
		return fmt.Errorf("ethereum address must be 20 bytes")
	}
	copy((*b)[:], u)
	return err
}

func (b *addressType) Compare(v interface{}) bool {
	if v, ok := v.(*addressType); ok {
		return *b == *v
	}
	return false
}

// addressType marshals/unmarshals as hash.
type hashType [hashLength]byte

func newHash(hash string) hashType {
	h := hashType{}
	b, err := hexToBytes([]byte(hash))
	if err != nil {
		return h
	}
	if len(b) != hashLength {
		return h
	}
	copy(h[:], b)
	return h
}

// MarshalJSON implements json.Marshaler.
func (b hashType) MarshalJSON() ([]byte, error) {
	return naiveQuote(bytesToHex(b[:])), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *hashType) UnmarshalJSON(input []byte) error {
	u, err := hexToBytes(naiveUnquote(input))
	if len(u) != hashLength {
		return fmt.Errorf("hash must be 32 bytes")
	}
	copy((*b)[:], u)
	return err
}

func (b *hashType) Compare(v interface{}) bool {
	if v, ok := v.(*hashType); ok {
		return *b == *v
	}
	return false
}

type blockType struct {
	Number           numberType  `json:"number"`
	Hash             hashType    `json:"hash"`
	ParentHash       hashType    `json:"parentHash"`
	Nonce            numberType  `json:"nonce"`
	Sha3Uncles       hashType    `json:"sha3Uncles"`
	LogsBloom        bytesType   `json:"logsBloom"`
	TransactionsRoot hashType    `json:"transactionsRoot"`
	StateRoot        hashType    `json:"stateRoot"`
	ReceiptsRoot     hashType    `json:"receiptsRoot"`
	Miner            addressType `json:"miner"`
	MixHash          hashType    `json:"mixHash"`
	Difficulty       numberType  `json:"difficulty"`
	TotalDifficulty  numberType  `json:"totalDifficulty"`
	ExtraData        bytesType   `json:"extraData"`
	Size             numberType  `json:"size"`
	GasLimit         numberType  `json:"gasLimit"`
	GasUsed          numberType  `json:"gasUsed"`
	Timestamp        numberType  `json:"timestamp"`
	Uncles           []hashType  `json:"uncles"`
}

type blockTxHashesType struct {
	blockType
	Transactions []hashType `json:"transactions"`
}

type blockTxObjectsType struct {
	blockType
	Transactions []transactionType `json:"transactions"`
}

type transactionType struct {
	Hash             hashType    `json:"hash"`
	BlockHash        hashType    `json:"blockHash"`
	BlockNumber      numberType  `json:"blockNumber"`
	TransactionIndex numberType  `json:"transactionIndex"`
	From             addressType `json:"from"`
	To               addressType `json:"to"`
	Gas              numberType  `json:"gas"`
	GasPrice         numberType  `json:"gasPrice"`
	Input            bytesType   `json:"input"`
	Nonce            numberType  `json:"nonce"`
	Value            numberType  `json:"value"`
	V                numberType  `json:"v"`
	R                hashType    `json:"r"`
	S                hashType    `json:"s"`
}

type logType struct {
	Address          addressType `json:"address"`
	Topics           []hashType  `json:"topics"`
	Data             bytesType   `json:"data"`
	BlockHash        hashType    `json:"blockHash"`
	BlockNumber      numberType  `json:"blockNumber"`
	TransactionHash  hashType    `json:"transactionHash"`
	TransactionIndex numberType  `json:"transactionIndex"`
	LogIndex         numberType  `json:"logIndex"`
	Removed          bool        `json:"removed"`
}

type transactionReceiptType struct {
	TransactionHash   hashType     `json:"transactionHash"`
	TransactionIndex  numberType   `json:"transactionIndex"`
	BlockHash         hashType     `json:"blockHash"`
	BlockNumber       numberType   `json:"blockNumber"`
	From              addressType  `json:"from"`
	To                addressType  `json:"to"`
	CumulativeGasUsed numberType   `json:"cumulativeGasUsed"`
	GasUsed           numberType   `json:"gasUsed"`
	ContractAddress   *addressType `json:"contractAddress"`
	Logs              []logType    `json:"logs"`
	LogsBloom         bytesType    `json:"logsBloom"`
	Root              *hashType    `json:"root"`
	Status            *numberType  `json:"status"`
}

type feeHistoryType struct {
	OldestBlock   numberType     `json:"oldestBlock"`
	Reward        [][]numberType `json:"reward"`
	BaseFeePerGas []numberType   `json:"baseFeePerGas"`
	GasUsedRatio  []float64      `json:"gasUsedRatio"`
}

func bigIntToHex(u *big.Int) []byte {
	r := make([]byte, 2, 10)
	copy(r, `0x`)
	r = u.Append(r, 16)
	return r
}

func hexToBigInt(h []byte) (*big.Int, error) {
	if has0xPrefix(h) {
		h = h[2:]
	}
	i, ok := new(big.Int).SetString(string(h), 16)
	if !ok {
		return nil, errors.New("invalid hex string")
	}
	return i, nil
}

func bytesToHex(b []byte) []byte {
	r := make([]byte, len(b)*2+2)
	copy(r, `0x`)
	hex.Encode(r[2:], b)
	return r
}

func hexToBytes(h []byte) ([]byte, error) {
	if has0xPrefix(h) {
		h = h[2:]
	}
	r := make([]byte, len(h)/2)
	_, err := hex.Decode(r, h)
	return r, err
}

func has0xPrefix(i []byte) bool {
	return len(i) >= 2 && i[0] == '0' && (i[1] == 'x' || i[1] == 'X')
}

func naiveQuote(i []byte) []byte {
	b := make([]byte, len(i)+2)
	b[0] = '"'
	b[len(b)-1] = '"'
	copy(b[1:], i)
	return b
}

func naiveUnquote(i []byte) []byte {
	if len(i) >= 2 && i[0] == '"' && i[len(i)-1] == '"' {
		return i[1 : len(i)-1]
	}
	return i
}
