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
	"fmt"
	"os"

	suite "github.com/makerdao/oracle-suite"
	"github.com/makerdao/oracle-suite/internal/gofer/marshal"
)

// exitCode to be returned by the application.
var exitCode = 0

func main() {
	opts := options{
		Format:  formatTypeValue{format: marshal.NDJSON},
		Version: suite.Version,
	}

	rootCmd := NewRootCommand(&opts)
	rootCmd.AddCommand(
		NewPairsCmd(&opts),
		NewPricesCmd(&opts),
		NewAgentCmd(&opts),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %s\n", err)
		if exitCode == 0 {
			os.Exit(1)
		}
	}
	os.Exit(exitCode)
}
