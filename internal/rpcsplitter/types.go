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

// MarshalJSON returns m as the JSON encoding of m.
func (t jsonType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.j)
}

// UnmarshalJSON sets *m to a copy of data.
func (t *jsonType) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &t.j)
}

func (t *jsonType) Compare(v interface{}) bool {
	switch vt := v.(type) {
	case *jsonType:
		return compare(t.j, vt.j)
	}
	return false
}

type blockNumberType big.Int

const earliestBlockNumber = -1
const latestBlockNumber = -2
const pendingBlockNumber = -3

// MarshalJSON implements json.Marshaler.
func (n blockNumberType) MarshalJSON() ([]byte, error) {
	return naiveQuote(bigIntToHex((*big.Int)(&n))), nil
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
		if u.Cmp(big.NewInt(math.MaxInt64)) > 0 {
			return fmt.Errorf("block number larger than int64")
		}
		*n = blockNumberType(*u)
		return err
	}
}

func (n *blockNumberType) Compare(v interface{}) bool {
	switch vt := v.(type) {
	case *blockNumberType:
		return n.Big().Cmp(vt.Big()) == 0
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
	switch vt := v.(type) {
	case *numberType:
		return n.Big().Cmp(vt.Big()) == 0
	}
	return false
}

func (n *numberType) Big() *big.Int {
	return (*big.Int)(n)
}

// bytesType marshals/unmarshals as a JSON string with 0x prefix.
// The empty slice marshals as "0x".
type bytesType []byte

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
	switch vt := v.(type) {
	case *bytesType:
		return bytes.Equal(*b, *vt)
	}
	return false
}

// addressType marshals/unmarshals as an Ethereum address.
type addressType [addressLength]byte

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
	switch vt := v.(type) {
	case *addressType:
		return *b == *vt
	}
	return false
}

// addressType marshals/unmarshals as hash.
type hashType [hashLength]byte

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
	switch vt := v.(type) {
	case *hashType:
		return *b == *vt
	}
	return false
}

type blockType struct {
	Number           numberType        `json:"number"`
	Hash             hashType          `json:"hash"`
	ParentHash       hashType          `json:"parentHash"`
	Nonce            numberType        `json:"nonce"`
	Sha3Uncles       hashType          `json:"sha3Uncles"`
	LogsBloom        bytesType         `json:"logsBloom"`
	TransactionsRoot hashType          `json:"transactionsRoot"`
	StateRoot        hashType          `json:"stateRoot"`
	ReceiptsRoot     hashType          `json:"receiptsRoot"`
	Miner            addressType       `json:"miner"`
	MixHash          hashType          `json:"mixHash"`
	Difficulty       numberType        `json:"difficulty"`
	TotalDifficulty  numberType        `json:"totalDifficulty"`
	ExtraData        bytesType         `json:"extraData"`
	Size             numberType        `json:"size"`
	GasLimit         numberType        `json:"gasLimit"`
	GasUsed          numberType        `json:"gasUsed"`
	Timestamp        numberType        `json:"timestamp"`
	Transactions     []transactionType `json:"transactions"`
	Uncles           []hashType        `json:"uncles"`
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
