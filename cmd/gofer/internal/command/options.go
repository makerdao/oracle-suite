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

package command

import (
	"fmt"
	"strings"

	"github.com/makerdao/gofer/cmd/gofer/internal/marshal"
)

// These are the command options that can be set by CLI flags.
type Options struct {
	ConfigFilePath string
	OutputFormat   FormatTypeValue
}

var formatMap = map[marshal.FormatType]string{
	marshal.Plain:  "plain",
	marshal.Trace:  "trace",
	marshal.JSON:   "json",
	marshal.NDJSON: "ndjson",
}

// FormatTypeValue is a wrapper for the FormatType to allow implement
// the flag.Value and spf13.pflag.Value interfaces.
type FormatTypeValue struct {
	Format marshal.FormatType
}

// Will return the default value if none is set and will fail if the `format` is set to an unsupported value for some reason.
func (v *FormatTypeValue) String() string {
	if v != nil {
		return formatMap[v.Format]
	}
	return formatMap[marshal.Plain]
}

func (v *FormatTypeValue) Set(s string) error {
	s = strings.ToLower(s)

	for ct, st := range formatMap {
		if s == st {
			v.Format = ct
			return nil
		}
	}

	return fmt.Errorf("unsupported format")
}

func (v *FormatTypeValue) Type() string {
	return "plain|trace|json|ndjson"
}
