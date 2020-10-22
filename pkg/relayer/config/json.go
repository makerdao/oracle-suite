package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/makerdao/gofer/internal/ethereum"
	"github.com/makerdao/gofer/pkg/relayer"
)

type JSON struct {
	Ethereum JSONEthereum        `json:"ethereum"`
	Feeds    []string            `json:"feeds"`   // TODO
	Options  JSONOptions         `json:"options"` // TODO
	Pairs    map[string]JSONPair `json:"pairs"`
}

type JSONEthereum struct {
	From     string `json:"from"`
	Keystore string `json:"keystore"`
	Password string `json:"password"`
	RPC      string `json:"rpc"`
}

type JSONOptions struct {
	Interval int  `json:"interval"` // TODO
	MsgLimit int  `json:"msgLimit"` // TODO
	Verbose  bool `json:"verbose"`  // TODO
}

type JSONPair struct {
	Oracle           string  `json:"oracle"`
	OracleSpread     float64 `json:"oracleSpread"`
	OracleExpiration int64   `json:"oracleExpiration"`
	MsgExpiration    int64   `json:"msgExpiration"`
}

type JSONConfigErr struct {
	Err error
}

func (e JSONConfigErr) Error() string {
	return e.Err.Error()
}

func ParseJSONFile(path string) (*JSON, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load JSON config file: %w", err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, JSONConfigErr{fmt.Errorf("failed to load JSON config file: %w", err)}
	}

	return ParseJSON(b)
}

func ParseJSON(b []byte) (*JSON, error) {
	j := &JSON{}
	err := json.Unmarshal(b, j)
	if err != nil {
		return nil, JSONConfigErr{err}
	}

	return j, nil
}

func (j *JSON) MakeRelayer() (*relayer.Relayer, error) {
	client, err := ethclient.Dial(j.Ethereum.RPC)
	if err != nil {
		return nil, err
	}

	wallet, err := ethereum.NewWallet(
		j.Ethereum.Keystore,
		j.Ethereum.Password,
		common.HexToAddress(j.Ethereum.From),
	)
	if err != nil {
		return nil, err
	}

	eth := ethereum.NewClient(client, wallet)
	rel := relayer.NewRelayer(eth, wallet)

	for name, pair := range j.Pairs {
		rel.AddPair(relayer.Pair{
			AssetPair:        name,
			Oracle:           common.HexToAddress(pair.Oracle),
			OracleSpread:     pair.OracleSpread,
			OracleExpiration: time.Second * time.Duration(pair.OracleExpiration),
			MsgExpiration:    time.Second * time.Duration(pair.MsgExpiration),
		})
	}

	return rel, nil
}
