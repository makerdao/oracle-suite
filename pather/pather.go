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

package pather

import (
	"makerdao/gofer/model"
)

// Pather describes a set of asset pairs and how to trade between them
type Pather interface {
	// Pairs returns a list of Pairs that are tradeable
	Pairs() []*model.Pair
	// Path returns PricePaths describing how to trade between two assets
	Path(*model.Pair) *model.PricePaths
}

// FilterPotentialPricePoints returns the PotentialPricePoints that are required
// to complete the PricePaths given and nil if path is not possible to complete
// with the given PotentialPricePoints
func FilterPotentialPricePoints(ppath *model.PricePaths, ppps []*model.PotentialPricePoint) []*model.PotentialPricePoint {
	resIndex := make(map[*model.PotentialPricePoint]bool)
	for _, path := range ppath.Paths {
		index := make(map[model.Pair]bool)
		for _, pair := range path {
			index[*pair] = true
		}

		pppIndex := make(map[*model.PotentialPricePoint]bool)
		pairIndex := make(map[model.Pair]bool)
		for _, ppp := range ppps {
			if _, ok := index[*ppp.Pair]; ok {
				pppIndex[ppp] = true
				pairIndex[*ppp.Pair] = true
			}
		}

		if len(pairIndex) == len(index) {
			for ppp := range pppIndex {
				resIndex[ppp] = true
			}
		}
	}

	var result []*model.PotentialPricePoint
	for ppp := range resIndex {
		result = append(result, ppp)
	}
	return result
}
