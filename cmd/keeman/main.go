//  Copyright (C) 2021 Maker Ecosystem Growth Holdings, INC.
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
	"log"
	"os"
)

func main() {
	if err := cmd(os.Args[1:]); err != nil {
		log.Fatalln(err)
	}
}

func cmd(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing command")
	}

	c := args[0]
	switch c {
	case "gen":
		return cmdGen(args)
	case "der":
		return cmdDer(args)
	}

	return fmt.Errorf("unknown command: %s", c)
}
