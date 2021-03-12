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
	"github.com/makerdao/gofer/pkg/gofer"
	"github.com/makerdao/gofer/pkg/gofer/graph/feeder"
	"github.com/makerdao/gofer/pkg/gofer/graph/nodes"
)

// AsyncGraph implements the gofer.Gofer interface. It works just like Graph
// but allows to update prices asynchronously.
type AsyncGraph struct {
	*Graph
	feeder *feeder.Feeder
}

// NewGraph returns a new AsyncGraph instance.
func NewAsyncGraph(g map[gofer.Pair]nodes.Aggregator, f *feeder.Feeder) *AsyncGraph {
	return &AsyncGraph{
		Graph:  NewGraph(g, nil),
		feeder: f,
	}
}

// Start starts asynchronous price updater.
func (a *AsyncGraph) Start() error {
	ns, _ := a.findNodes()
	return a.feeder.Start(ns...)
}

// Start stops asynchronous price updater.
func (a *AsyncGraph) Stop() error {
	a.feeder.Stop()
	return nil
}
