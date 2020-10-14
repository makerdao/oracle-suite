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
	"log"
	"time"

	"github.com/makerdao/gofer/pkg/graph"
)

func Feed(feeder *graph.Feeder, nodes []graph.Node) {
	if err := feeder.Feed(nodes); err != nil {
		log.Println(err)
	}
}

func ScheduleFeeding(feeder *graph.Feeder, nodes []graph.Node) func() {
	done := make(chan bool)
	ticker := time.NewTicker(getMinTTL(nodes))
	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
				return
			case <-ticker.C:
				log.Println("repopulating data graph")
				Feed(feeder, nodes)
			}
		}
	}()
	return func() {
		done <- true
	}
}

func getMinTTL(nodes []graph.Node) time.Duration {
	minTTL := time.Duration(0)
	graph.Walk(func(node graph.Node) {
		if feedable, ok := node.(graph.Feedable); ok {
			if minTTL == 0 || feedable.MinTTL() < minTTL {
				minTTL = feedable.MinTTL()
			}
		}
	}, nodes...)

	if minTTL < time.Second {
		return time.Second
	}

	return minTTL
}
