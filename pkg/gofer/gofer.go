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

package gofer

import (
	"fmt"
	"strings"
	"time"
)

// Pair represents an asset pair.
type Pair struct {
	Base  string
	Quote string
}

func NewPair(s string) (Pair, error) {
	ss := strings.Split(s, "/")
	if len(ss) != 2 {
		return Pair{}, fmt.Errorf("couldn't parse pair \"%s\"", s)
	}
	return Pair{Base: strings.ToUpper(ss[0]), Quote: strings.ToUpper(ss[1])}, nil
}

func NewPairs(s ...string) ([]Pair, error) {
	var r []Pair
	for _, p := range s {
		pr, err := NewPair(p)
		if err != nil {
			return nil, err
		}
		r = append(r, pr)
	}
	return r, nil
}

func (p Pair) Empty() bool {
	return p.Base == "" && p.Quote == ""
}

func (p Pair) Equal(c Pair) bool {
	return p.Base == c.Base && p.Quote == c.Quote
}

func (p Pair) String() string {
	return fmt.Sprintf("%s/%s", p.Base, p.Quote)
}

// Node is a simplified representation of a graph structure which is used to
// calculate asset pair prices. The main purpose of this structure is to help
// the end user understand how prices are derived.
//
// This structure may look different depending on the gofer.Gofer
// implementation.
type Node struct {
	// Type is used to differentiate between node types.
	Type string
	// Parameters is a optional list of node's parameters.
	Parameters map[string]string
	// Pair is a asset pair for which this node returns a tick price.
	Pair Pair
	// Children is a list of children nodes used to calculate price. It is
	// empty if the node provides the asset pair price directly.
	Children []*Node
}

// Tick represents tick price for single pair. If the tick price was calculated
// indirectly it will also contain all ticks used to calculate prices.
type Tick struct {
	Type       string
	Parameters map[string]string
	Pair       Pair
	Price      float64
	Bid        float64
	Ask        float64
	Volume24h  float64
	Time       time.Time
	Ticks      []*Tick
	Error      string
}

type Gofer interface {
	// Nodes returns nodes for given pairs. If no pairs are specified, nodes
	// for all known pairs will be returned.
	Nodes(pairs ...Pair) (map[Pair]*Node, error)
	// Tick returns fresh tick price for the given pair.
	Tick(pair Pair) (*Tick, error)
	// Ticks returns fresh tick prices for the given pairs. If no pairs are
	// specified, ticks for all known pairs will be returned.
	Ticks(pairs ...Pair) (map[Pair]*Tick, error)
	// Pairs returns all known pairs.
	Pairs() ([]Pair, error)
}

// StartableGofer interface represents Gofer instances that have to be started
// first to work properly.
type StartableGofer interface {
	Gofer
	Start() error
	Stop() error
}
