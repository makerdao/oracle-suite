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
	"github.com/sirupsen/logrus"
)

const defaultVerbosity = logrus.WarnLevel

type Verbosity struct {
	wasSet    bool
	verbosity logrus.Level
}

// String implements the pflag.Value interface.
func (f *Verbosity) String() string {
	if !f.wasSet {
		return defaultVerbosity.String()
	}
	return f.verbosity.String()
}

// Set implements the pflag.Value interface.
func (f *Verbosity) Set(v string) (err error) {
	f.verbosity, err = logrus.ParseLevel(v)
	if err != nil {
		return err
	}
	f.wasSet = true
	return err
}

// Type implements the pflag.Value interface.
func (f *Verbosity) Type() string {
	var s string
	for _, l := range logrus.AllLevels {
		if len(s) > 0 {
			s += "|"
		}
		s += l.String()
	}
	return s
}

func (f *Verbosity) Level() logrus.Level {
	if f.verbosity == 0 {
		return defaultVerbosity
	}
	return f.verbosity
}
