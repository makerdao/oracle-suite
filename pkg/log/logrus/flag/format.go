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

package flag

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/makerdao/oracle-suite/pkg/log/logrus/formatter"
)

// formattersMap is a map of supported logrus formatters. It is safe to add
// custom formatters to this map.
var formattersMap = map[string]func() logrus.Formatter{
	"text": func() logrus.Formatter {
		return &formatter.XFilterFormatter{Formatter: &logrus.TextFormatter{}}
	},
	"json": func() logrus.Formatter {
		return &formatter.JSONFormatter{}
	},
}

const defaultFormatter = "text"

// Format implements pflag.Value. It represents a flag that allow
// to choose a different logrus formatter.
type Format struct {
	format string
}

// String implements the pflag.Value interface.
func (f *Format) String() string {
	if f.format == "" {
		return defaultFormatter
	}
	return f.format
}

// Set implements the pflag.Value interface.
func (f *Format) Set(v string) error {
	v = strings.ToLower(v)
	if _, ok := formattersMap[v]; !ok {
		return fmt.Errorf("unsupported format")
	}
	f.format = v
	return nil
}

// Type implements the pflag.Value interface.
func (f *Format) Type() string {
	return "text|json"
}

// Formatter returns the logrus.Formatter for selected type.
func (f *Format) Formatter() logrus.Formatter {
	if f.format == "" {
		return formattersMap[defaultFormatter]()
	}
	return formattersMap[f.format]()
}
