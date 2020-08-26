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

package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/makerdao/gofer/aggregator"
)

// Config holds CLI's config options for immediate parsing
type Config struct {
	Aggregator struct {
		Name       string
		Parameters struct {
			PriceModels aggregator.PriceModelMap `json:"pricemodels"`
		}
	}
}

func ReadConfig(path string) (*Config, error) {
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

	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config file: %s %w", path, err)
	}

	return &c, nil
}
