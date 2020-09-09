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
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/makerdao/gofer/pkg/model"
)

type exchangeLister interface {
	Exchanges(pairs ...*model.Pair) []*model.Exchange
}

func Exchanges(args []string, l exchangeLister, m ReadWriteCloser) error {
	var pairs []*model.Pair
	for _, pair := range args {
		p, err := model.NewPairFromString(pair)
		if err != nil {
			return err
		}
		pairs = append(pairs, model.NewPair(p.Base, p.Quote))
	}

	exchanges := l.Exchanges(pairs...)

	sort.SliceStable(exchanges, func(i, j int) bool {
		return exchanges[i].Name < exchanges[j].Name
	})

	var er error
	for _, e := range exchanges {
		err := m.Write(e, er)
		if err != nil {
			return err
		}
	}

	err := m.Close()
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
