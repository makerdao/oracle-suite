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
	"github.com/makerdao/oracle-suite/pkg/gofer"
	"github.com/makerdao/oracle-suite/pkg/gofer/graph/feeder"
	"github.com/makerdao/oracle-suite/pkg/gofer/graph/nodes"
)

// AsyncGofer implements the gofer.Gofer interface. It works just like Graph
// but allows to update prices asynchronously.
type AsyncGofer struct {
	*Gofer
	feeder *feeder.Feeder
}

// NewAsyncGofer returns a new AsyncGofer instance.
func NewAsyncGofer(g map[gofer.Pair]nodes.Aggregator, f *feeder.Feeder) *AsyncGofer {
	return &AsyncGofer{
		Gofer:  NewGofer(g, nil),
		feeder: f,
	}
}

// Start starts asynchronous price updater.
func (a *AsyncGofer) Start() error {
	ns, _ := a.findNodes()
	return a.feeder.Start(ns...)
}

// Start stops asynchronous price updater.
func (a *AsyncGofer) Stop() error {
	a.feeder.Stop()
	return nil
}
