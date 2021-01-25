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

package json

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/makerdao/gofer/pkg/spectre/config"
)

type ConfigErr struct {
	Err error
}

func (e ConfigErr) Error() string {
	return e.Err.Error()
}

func ParseJSONFile(cfg *config.Config, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to load JSON config file: %w", err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return ConfigErr{fmt.Errorf("failed to load JSON config file: %w", err)}
	}

	return ParseJSON(cfg, b)
}

func ParseJSON(cfg *config.Config, b []byte) error {
	err := json.Unmarshal(b, cfg)
	if err != nil {
		return ConfigErr{err}
	}
	return nil
}
