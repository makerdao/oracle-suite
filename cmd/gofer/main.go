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
	"log"

	"github.com/makerdao/gofer/cmd/gofer/internal/command"
	"github.com/makerdao/gofer/cmd/gofer/internal/command/exchanges"
	"github.com/makerdao/gofer/cmd/gofer/internal/command/pairs"
	"github.com/makerdao/gofer/cmd/gofer/internal/command/price"
)

func main() {
	var opts command.Options
	rootCmd := command.New(&opts)
	rootCmd.AddCommand(exchanges.New(&opts), pairs.New(&opts), price.New(&opts))
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
