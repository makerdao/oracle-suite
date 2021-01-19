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

package null

import "github.com/makerdao/gofer/pkg/log"

type Null struct{}

func New() *Null {
	return &Null{}
}

func (n *Null) WithField(_ string, _ interface{}) log.Logger { return &Null{} }
func (n *Null) WithFields(_ log.Fields) log.Logger           { return &Null{} }
func (n *Null) WithError(_ error) log.Logger                 { return &Null{} }
func (n *Null) Debugf(_ string, _ ...interface{})            {}
func (n *Null) Infof(_ string, _ ...interface{})             {}
func (n *Null) Warnf(_ string, _ ...interface{})             {}
func (n *Null) Errorf(_ string, _ ...interface{})            {}
func (n *Null) Panicf(_ string, _ ...interface{})            {}
func (n *Null) Debug(_ ...interface{})                       {}
func (n *Null) Info(_ ...interface{})                        {}
func (n *Null) Warn(_ ...interface{})                        {}
func (n *Null) Error(_ ...interface{})                       {}
func (n *Null) Panic(_ ...interface{})                       {}
func (n *Null) Debugln(_ ...interface{})                     {}
func (n *Null) Infoln(_ ...interface{})                      {}
func (n *Null) Warnln(_ ...interface{})                      {}
func (n *Null) Errorln(_ ...interface{})                     {}
func (n *Null) Panicln(_ ...interface{})                     {}
