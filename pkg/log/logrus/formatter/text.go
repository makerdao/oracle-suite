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

package formatter

import (
	"strings"

	"github.com/sirupsen/logrus"
)

// TextFormatter is a wrapper for the original TextFormatter from the logrus
// package. It removes all fields with the "x-" prefix. This will allow to add
// more data fields to logs without making the CLI output to messy.
type TextFormatter struct {
	f logrus.TextFormatter
}

func (f *TextFormatter) Format(e *logrus.Entry) ([]byte, error) {
	data := logrus.Fields{}
	for k, v := range e.Data {
		if !strings.HasPrefix(k, "x-") {
			data[k] = v
		}
	}
	e.Data = data
	return f.f.Format(e)
}
