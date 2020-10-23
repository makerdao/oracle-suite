package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/makerdao/gofer/internal/ethereum"
	"github.com/makerdao/gofer/pkg/ghost"
	"github.com/makerdao/gofer/pkg/gofer"
	"github.com/makerdao/gofer/pkg/transport/serf"
)

type JSON struct {
	Ethereum JSONEthereum        `json:"ethereum"`
	Serf     JSONSerf            `json:"serf"`
	Options  JSONOptions         `json:"options"`
	Pairs    map[string]JSONPair `json:"pairs"`
}

type JSONEthereum struct {
	From     string `json:"from"`
	Keystore string `json:"keystore"`
	Password string `json:"password"`
}

type JSONSerf struct {
	RPC string `json:"rpc"`
}

type JSONOptions struct {
	Interval         int  `json:"interval"`
	MsgLimit         int  `json:"msgLimit"`
	SrcTimeout       int  `json:"srcTimeout"`
	GoferTimeout     int  `json:"goferTimeout"`
	GoferCacheExpiry int  `json:"goferCacheExpiry"`
	GoferMinMedian   int  `json:"goferMinMedian"`
	Verbose          bool `json:"verbose"`
}

type JSONPair struct {
	MsgExpiration int     `json:"msgExpiration"`
	MsgSpread     float64 `json:"msgSpread"`
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

func (j *JSON) MakeGhost(gofer *gofer.Gofer) (*ghost.Ghost, error) {
	wallet, err := ethereum.NewWallet(
		j.Ethereum.Keystore,
		j.Ethereum.Password,
		common.HexToAddress(j.Ethereum.From),
	)
	if err != nil {
		return nil, err
	}

	transport, err := serf.NewSerf(j.Serf.RPC)
	if err != nil {
		return nil, err
	}

	gho := ghost.NewGhost(gofer, wallet, transport, time.Second*time.Duration(j.Options.Interval))
	for name, pair := range j.Pairs {
		err := gho.AddPair(ghost.Pair{
			AssetPair:        name,
			OracleSpread:     pair.MsgSpread,
			OracleExpiration: time.Second * time.Duration(pair.MsgExpiration),
		})
		if err != nil {
			return nil, err
		}
	}

	return gho, nil
}
