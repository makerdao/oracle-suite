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

package logrus

import (
	"github.com/sirupsen/logrus"

	"github.com/makerdao/gofer/pkg/log"
)

type Logger struct {
	Logger logrus.FieldLogger
}

func New(logger logrus.FieldLogger) log.Logger {
	return &Logger{Logger: logger}
}

func (d *Logger) WithField(key string, value interface{}) log.Logger {
	return &Logger{Logger: d.Logger.WithField(key, value)}
}

func (d *Logger) WithFields(fields log.Fields) log.Logger {
	return &Logger{Logger: d.Logger.WithFields(fields)}
}

func (d *Logger) WithError(err error) log.Logger {
	if fErr, ok := err.(log.ErrorWithFields); ok {
		return &Logger{Logger: d.Logger.WithFields(fErr.Fields()).WithError(err)}
	}

	return &Logger{Logger: d.Logger.WithError(err)}
}

func (d *Logger) Debugf(format string, args ...interface{}) {
	d.Logger.Debugf(format, args...)
}

func (d *Logger) Infof(format string, args ...interface{}) {
	d.Logger.Infof(format, args...)
}

func (d *Logger) Warnf(format string, args ...interface{}) {
	d.Logger.Warnf(format, args...)
}

func (d *Logger) Errorf(format string, args ...interface{}) {
	d.Logger.Errorf(format, args...)
}

func (d *Logger) Fatalf(format string, args ...interface{}) {
	d.Logger.Fatalf(format, args...)
}

func (d *Logger) Panicf(format string, args ...interface{}) {
	d.Logger.Panicf(format, args...)
}

func (d *Logger) Debug(args ...interface{}) {
	d.Logger.Debug(args...)
}

func (d *Logger) Info(args ...interface{}) {
	d.Logger.Info(args...)
}

func (d *Logger) Warn(args ...interface{}) {
	d.Logger.Warn(args...)
}

func (d *Logger) Error(args ...interface{}) {
	d.Logger.Error(args...)
}

func (d *Logger) Fatal(args ...interface{}) {
	d.Logger.Fatal(args...)
}

func (d *Logger) Panic(args ...interface{}) {
	d.Logger.Panic(args...)
}

func (d *Logger) Debugln(args ...interface{}) {
	d.Logger.Debugln(args...)
}

func (d *Logger) Infoln(args ...interface{}) {
	d.Logger.Infoln(args...)
}

func (d *Logger) Warnln(args ...interface{}) {
	d.Logger.Warnln(args...)
}

func (d *Logger) Errorln(args ...interface{}) {
	d.Logger.Errorln(args...)
}

func (d *Logger) Fatalln(args ...interface{}) {
	d.Logger.Fatalln(args...)
}

func (d *Logger) Panicln(args ...interface{}) {
	d.Logger.Panicln(args...)
}
