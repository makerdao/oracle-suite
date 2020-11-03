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

package transport

type Status struct {
	Error error
}

type Event interface {
	PayloadMarshall() ([]byte, error)
	PayloadUnmarshall([]byte) error
}

type Transport interface {
	Broadcast(eventName string, payload Event) error
	Subscribe(eventName string) error
	Unsubscribe(eventName string) error
	WaitFor(eventName string, payload Event) chan Status
	Close() error
}