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

package graph

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
)

type IngestedIncompatiblePairErr struct {
	Given    Pair
	Expected Pair
}

func (e IngestedIncompatiblePairErr) Error() string {
	return fmt.Sprintf(
		"a tick with different pair ignested to the OriginNode, %s given but %s was expected",
		e.Given,
		e.Expected,
	)
}

type IngestedIncompatibleOriginErr struct {
	Given    string
	Expected string
}

func (e IngestedIncompatibleOriginErr) Error() string {
	return fmt.Sprintf(
		"a tick from different origin ignested to the OriginNode, %s given but %s was expected",
		e.Given,
		e.Expected,
	)
}

// OriginNode contains a Tick fetched directly from an origin.
type OriginNode struct {
	originPair OriginPair
	tick       OriginTick
}

func NewOriginNode(originPair OriginPair) *OriginNode {
	return &OriginNode{
		originPair: originPair,
	}
}

func (n *OriginNode) OriginPair() OriginPair {
	return n.originPair
}

func (n *OriginNode) Ingest(tick OriginTick) error {
	var err error
	if !tick.Pair.Equal(n.originPair.Pair) {
		err = multierror.Append(err, IngestedIncompatiblePairErr{
			Given:    tick.Pair,
			Expected: n.originPair.Pair,
		})
	}

	if tick.Origin != n.originPair.Origin {
		err = multierror.Append(err, IngestedIncompatibleOriginErr{
			Given:    tick.Origin,
			Expected: n.originPair.Origin,
		})
	}

	if err != nil {
		// TBD: Do we want to assign this error to OriginTick?
		n.tick = OriginTick{
			Tick:   Tick{},
			Origin: "",
			Error:  err,
		}

		return err
	}

	n.tick = tick

	return nil
}

func (n *OriginNode) Tick() OriginTick {
	return n.tick
}

func (n OriginNode) Children() []Node {
	return []Node{}
}
