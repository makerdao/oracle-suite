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

package log

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Fields = map[string]interface{}

type ErrorWithFields interface {
	error
	Fields() Fields
}

type Logger interface {
	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
	WithError(err error) Logger

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Panic(args ...interface{})

	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Warnln(args ...interface{})
	Errorln(args ...interface{})
	Panicln(args ...interface{})
}

func Format(s ...interface{}) []string {
	r := make([]string, len(s))
	for i, s := range s {
		switch ts := s.(type) {
		case error:
			r[i] = ts.Error()
		default:
			rtype := reflect.TypeOf(s)
			switch rtype.Kind() {
			case reflect.Struct, reflect.Interface, reflect.Map, reflect.Slice, reflect.Array:
				j, err := json.Marshal(s)
				if err != nil {
					r[i] = err.Error()
				} else {
					r[i] = string(j)
				}
			case reflect.Ptr:
				rvalue := reflect.ValueOf(s)
				r[i] = Format(rvalue.Elem().Interface())[0]
			default:
				r[i] = fmt.Sprint(s)
			}
		}
	}
	return r
}
