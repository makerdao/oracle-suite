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

import "github.com/sirupsen/logrus"

type Logger = logrus.FieldLogger
type Fields = map[string]interface{}

type FieldLogger struct {
	Logger Logger
	Fields Fields
}

type ErrorWithFields interface {
	error
	Fields() Fields
}

func WrapLogger(parent Logger, fields Fields) Logger {
	return &FieldLogger{
		Logger: parent,
		Fields: fields,
	}
}

func (d *FieldLogger) WithField(key string, value interface{}) *logrus.Entry {
	return d.Logger.WithFields(d.Fields).WithField(key, value)
}

func (d *FieldLogger) WithFields(fields logrus.Fields) *logrus.Entry {
	return d.Logger.WithFields(d.Fields).WithFields(fields)
}

func (d *FieldLogger) WithError(err error) *logrus.Entry {
	if fErr, ok := err.(ErrorWithFields); ok {
		return d.Logger.WithFields(d.Fields).WithFields(fErr.Fields()).WithError(err)
	}

	return d.Logger.WithFields(d.Fields).WithError(err)
}

func (d *FieldLogger) Debugf(format string, args ...interface{}) {
	d.Logger.WithFields(d.Fields).Debugf(format, args...)
}

func (d *FieldLogger) Infof(format string, args ...interface{}) {
	d.Logger.WithFields(d.Fields).Infof(format, args...)
}

func (d *FieldLogger) Printf(format string, args ...interface{}) {
	d.Logger.WithFields(d.Fields).Printf(format, args...)
}

func (d *FieldLogger) Warnf(format string, args ...interface{}) {
	d.Logger.WithFields(d.Fields).Warnf(format, args...)
}

func (d *FieldLogger) Warningf(format string, args ...interface{}) {
	d.Logger.WithFields(d.Fields).Warningf(format, args...)
}

func (d *FieldLogger) Errorf(format string, args ...interface{}) {
	d.Logger.WithFields(d.Fields).Errorf(format, args...)
}

func (d *FieldLogger) Fatalf(format string, args ...interface{}) {
	d.Logger.WithFields(d.Fields).Fatalf(format, args...)
}

func (d *FieldLogger) Panicf(format string, args ...interface{}) {
	d.Logger.WithFields(d.Fields).Panicf(format, args...)
}

func (d *FieldLogger) Debug(args ...interface{}) {
	d.Logger.WithFields(d.Fields).Debug(args...)
}

func (d *FieldLogger) Info(args ...interface{}) {
	d.Logger.WithFields(d.Fields).Info(args...)
}

func (d *FieldLogger) Print(args ...interface{}) {
	d.Logger.WithFields(d.Fields).Print(args...)
}

func (d *FieldLogger) Warn(args ...interface{}) {
	d.Logger.WithFields(d.Fields).Warn(args...)
}

func (d *FieldLogger) Warning(args ...interface{}) {
	d.Logger.WithFields(d.Fields).Warning(args...)
}

func (d *FieldLogger) Error(args ...interface{}) {
	d.Logger.WithFields(d.Fields).Error(args...)
}

func (d *FieldLogger) Fatal(args ...interface{}) {
	d.Logger.WithFields(d.Fields).Fatal(args...)
}

func (d *FieldLogger) Panic(args ...interface{}) {
	d.Logger.WithFields(d.Fields).Panic(args...)
}

func (d *FieldLogger) Debugln(args ...interface{}) {
	d.Logger.WithFields(d.Fields).Debugln(args...)
}

func (d *FieldLogger) Infoln(args ...interface{}) {
	d.Logger.WithFields(d.Fields).Infoln(args...)
}

func (d *FieldLogger) Println(args ...interface{}) {
	d.Logger.WithFields(d.Fields).Println(args...)
}

func (d *FieldLogger) Warnln(args ...interface{}) {
	d.Logger.WithFields(d.Fields).Warnln(args...)
}

func (d *FieldLogger) Warningln(args ...interface{}) {
	d.Logger.WithFields(d.Fields).Warningln(args...)
}

func (d *FieldLogger) Errorln(args ...interface{}) {
	d.Logger.WithFields(d.Fields).Errorln(args...)
}

func (d *FieldLogger) Fatalln(args ...interface{}) {
	d.Logger.WithFields(d.Fields).Fatalln(args...)
}

func (d *FieldLogger) Panicln(args ...interface{}) {
	d.Logger.WithFields(d.Fields).Panicln(args...)
}
