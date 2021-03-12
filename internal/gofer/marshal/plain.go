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

package marshal

import (
	"bytes"
	"fmt"

	"github.com/makerdao/gofer/pkg/gofer"
)

type plain struct {
	items [][]byte
}

func newPlain() *plain {
	return &plain{}
}

// Read implements the Marshaller interface.
func (p *plain) Bytes() ([]byte, error) {
	return append(bytes.Join(p.items, []byte("\n")), '\n'), nil
}

// Write implements the Marshaller interface.
func (p *plain) Write(item interface{}) error {
	var i []byte
	switch typedItem := item.(type) {
	case *gofer.Price:
		i = p.handlePrice(typedItem)
	case *gofer.Model:
		i = p.handleNode(typedItem)
	default:
		return fmt.Errorf("unsupported data type")
	}

	p.items = append(p.items, i)
	return nil
}

func (*plain) handlePrice(price *gofer.Price) []byte {
	if price.Error != "" {
		return []byte(fmt.Sprintf("%s - %s", price.Pair, price.Error))
	}
	return []byte(fmt.Sprintf("%s %f", price.Pair, price.Price))
}

func (*plain) handleNode(node *gofer.Model) []byte {
	return []byte(node.Pair.String())
}
