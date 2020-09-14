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

func (n *OriginNode) Feed(tick OriginTick) {
	n.tick = tick
}

func (n *OriginNode) Tick() OriginTick {
	return n.tick
}

func (n OriginNode) Children() []Node {
	return []Node{}
}
