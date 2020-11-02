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

package events

import (
	"crypto/rand"
	"testing"
)

func Test_compress(t *testing.T) {
	data := make([]byte, 1024 * 1024 * 2) // 2MB
	_, _ = rand.Read(data)

	compressed, _ := compress(data)
	decompressed, _ := decompress(compressed)

	if len(data) != len(decompressed) {
		t.Error("decompressed data has different length")
	}

	for i, _ := range data {
		if data[i] != decompressed[i] {
			t.Error("decompressed data is not the same")
		}
	}
}
