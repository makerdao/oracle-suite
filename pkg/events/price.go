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

package events

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io/ioutil"

	"github.com/makerdao/gofer/internal/oracle"
)

var PriceEventName = "price"

type Price struct {
	Price *oracle.Price   `json:"price"`
	Trace json.RawMessage `json:"trace"`
}

func (p *Price) PayloadMarshall() ([]byte, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	return compress(b)
}

func (p *Price) PayloadUnmarshall(b []byte) error {
	b, err := decompress(b)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, p)
}

func compress(in []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(in); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
func decompress(in []byte) ([]byte, error) {
	var b bytes.Buffer
	b.Write(in)
	gz, err := gzip.NewReader(&b)
	if err != nil {
		return nil, err
	}
	out, err := ioutil.ReadAll(gz)
	if err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return out, nil
}
