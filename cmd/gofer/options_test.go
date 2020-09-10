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

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatTypeValue(t *testing.T) {
	for ct, st := range formatMap {
		t.Run(st, func(t *testing.T) {
			ftv := formatTypeValue{}
			err := ftv.Set(st)

			assert.Nil(t, err)
			assert.Equal(t, st, ftv.String())
			assert.Equal(t, ct, ftv.format)
		})
	}
}
