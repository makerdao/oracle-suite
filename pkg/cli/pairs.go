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

package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"github.com/makerdao/gofer/pkg/aggregator"
)

func Pairs(configFilePath string, m ReadWriteCloser) error {
	conf, err := readConfig(configFilePath)
	if err != nil {
		return err
	}

	pairs := make([]aggregator.Pair, 0)
	for k := range conf.Aggregator.Parameters.PriceModels {
		pairs = append(pairs, k)
	}

	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].String() < pairs[j].String()
	})

	for _, p := range pairs {
		err = m.Write(aggregator.PriceModelMap{p: conf.Aggregator.Parameters.PriceModels[p]}, nil)
		if err != nil {
			return err
		}
	}

	err = m.Close()
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(m)
	if err != nil {
		return err
	}

	fmt.Print(string(b))

	return nil
}

type config struct {
	Aggregator struct {
		Name       string
		Parameters struct {
			PriceModels aggregator.PriceModelMap `json:"pricemodels"`
		}
	}
}

func readConfig(path string) (*config, error) {
	cf, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %s %w", path, err)
	}
	defer func() {
		if err = cf.Close(); err != nil {
			log.Printf("error closing config file: %s %s", path, err)
		}
	}()

	b, err := ioutil.ReadAll(cf)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %s %w", path, err)
	}

	var c config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config file: %s %w", path, err)
	}

	return &c, nil
}
