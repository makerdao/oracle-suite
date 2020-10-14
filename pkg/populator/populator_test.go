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

package populator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/pkg/graph"
)

func Test_getMinTTL(t *testing.T) {
	p := graph.Pair{Base: "A", Quote: "B"}
	root := graph.NewMedianAggregatorNode(p, 1)
	ttl := time.Second * time.Duration(time.Now().Unix() + 10)
	on1 := graph.NewOriginNode(graph.OriginPair{Origin: "a", Pair: p}, 12 * time.Second, ttl)
	on2 := graph.NewOriginNode(graph.OriginPair{Origin: "b", Pair: p}, 5 * time.Second, ttl)
	on3 := graph.NewOriginNode(graph.OriginPair{Origin: "b", Pair: p}, 10 * time.Second, ttl)

	root.AddChild(on1)
	root.AddChild(on2)
	root.AddChild(on3)

	assert.Equal(t, 5 * time.Second, getMinTTL([]graph.Node{root}))
}

func Test_getMinTTL_SorterThanOneSecond(t *testing.T) {
	p := graph.Pair{Base: "A", Quote: "B"}
	root := graph.NewMedianAggregatorNode(p, 1)
	ttl := time.Second * time.Duration(time.Now().Unix() + 10)
	on1 := graph.NewOriginNode(graph.OriginPair{Origin: "a", Pair: p}, 12 * time.Second, ttl)
	on2 := graph.NewOriginNode(graph.OriginPair{Origin: "b", Pair: p}, -5 * time.Second, ttl)
	on3 := graph.NewOriginNode(graph.OriginPair{Origin: "b", Pair: p}, 0 * time.Second, ttl)

	root.AddChild(on1)
	root.AddChild(on2)
	root.AddChild(on3)

	assert.Equal(t, 1 * time.Second, getMinTTL([]graph.Node{root}))
}
