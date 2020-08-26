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

package marshal

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

// Marshaller is the interface which must be implemented by different
// marshallers used to format output for the CLI.
//
// The Read method implements the io.Reader interface. This method locks
// current goroutine until new data appear or the Close method is called.
//
// The Write method allows to add data asynchronously to the Marshaller. It
// allows to add new data, even if the previous ones have not been read. When
// Marshaller is closed it will panic.
type Marshaller interface {
	io.ReadCloser

	Write(item interface{}, err error) error
}

// Marshal implements the Marshaller interface. It wraps other marshaller based
// on argument passed to the NewMarshal method.
type Marshal struct {
	marshaller Marshaller
}

const (
	Plain FormatType = iota
	JSON
	NDJSON
	Trace
)

// FormatType describes format used by Marshal.
type FormatType int

// NewMarshal returns new Marshal instance.
func NewMarshal(format FormatType) (*Marshal, error) {
	switch format {
	case Plain:
		return &Marshal{marshaller: newPlain()}, nil
	case JSON:
		return &Marshal{marshaller: newJSON(false)}, nil
	case NDJSON:
		return &Marshal{marshaller: newJSON(true)}, nil
	case Trace:
		return &Marshal{marshaller: newTrace()}, nil
	}

	return nil, fmt.Errorf("unsupported format")
}

// Write implements the Marshaller interface.
func (m *Marshal) Write(aggregate interface{}, err error) error {
	return m.marshaller.Write(aggregate, err)
}

// Read implements the Marshaller interface.
func (m *Marshal) Read(p []byte) (n int, err error) {
	return m.marshaller.Read(p)
}

// Close implements the Marshaller interface.
func (m *Marshal) Close() error {
	return m.marshaller.Close()
}

// marshalledItem is the alias for []byte. It represent data that has been
// already marshalled.
type marshalledItem []byte

// String implements the fmt.Stringer interface.
func (m marshalledItem) String() string {
	return string(m)
}

// MarshalJSON implements the encode.Marshaler interface.
func (m marshalledItem) MarshalJSON() ([]byte, error) {
	return m, nil
}

type marshallerFunc func(interface{}, error) ([]marshalledItem, error)

// bufferedMarshaller helps to implement Marshaller interface.
type bufferedMarshaller struct {
	marshaller marshallerFunc
	live       bool
	items      []marshalledItem
	buffer     bytes.Buffer
	cond       *sync.Cond
	closed     bool
}

// newBufferedMarshaller returns new instance of bufferedMarshaller.
//
// The live flag indicates if data will be returned by the Read method
// immediately or all at once after Close method is called.
//
// The marshaller function contains the code used to convert objects passed
// to the Write method to bytes. If the live flag is set to false, this method
// will be called one more time to marshall slice of all previously marshalled
// objects. This last call can be detected by asserting passed argument to
// []marshalledItem.
func newBufferedMarshaller(live bool, marshaller marshallerFunc) *bufferedMarshaller {
	return &bufferedMarshaller{
		marshaller: marshaller,
		live:       live,
		items:      []marshalledItem{},
		buffer:     bytes.Buffer{},
		cond:       sync.NewCond(&sync.Mutex{}),
		closed:     false,
	}
}

// Implements the Marshaller interface.
func (b *bufferedMarshaller) Read(p []byte) (int, error) {
	if b.live {
		// When the live flag is set to true, data are returned as soon
		// as they appear:
		for {
			b.cond.L.Lock()

			// Push all items into the bytes buffer. This allows to pass
			// reading logic to bytes.Buffer:
			for _, item := range b.items {
				b.buffer.Write(item)
			}
			b.items = []marshalledItem{}

			// Read data from the bytes buffer:
			n, err := b.buffer.Read(p)

			if n == 0 && !b.closed {
				// Block reader until new data appear:
				b.cond.Wait()
				b.cond.L.Unlock()
				continue
			} else {
				b.cond.L.Unlock()
				return n, err
			}
		}
	} else {
		// When the live flag is set to false, data are returned after the
		// Close method is invoked:
		for {
			b.cond.L.Lock()

			if b.closed {
				if b.items != nil {
					// As above, all items are pushed into the bytes
					// buffer, then we set b.items to nil to indicate that
					// data are already marshaled. Otherwise code below will
					// invoke marshaller on every read.
					bs, err := b.marshaller(b.items, nil)
					if err != nil {
						return 0, err
					}
					for _, bt := range bs {
						b.buffer.Write(bt)
					}
					b.items = nil
				}

				b.cond.L.Unlock()
				return b.buffer.Read(p)
			}
			b.cond.Wait()
			b.cond.L.Unlock()
		}
	}
}

// Implements the Marshaller interface.
func (b *bufferedMarshaller) Write(item interface{}, err error) error {
	b.cond.L.Lock()
	defer func() {
		b.cond.Broadcast()
		b.cond.L.Unlock()
	}()

	if b.closed {
		return fmt.Errorf("unable to write to closed writer")
	}

	bs, err := b.marshaller(item, err)
	if err != nil {
		return err
	}

	b.items = append(b.items, bs...)

	return nil
}

// Implements the Marshaller interface.
func (b *bufferedMarshaller) Close() error {
	b.cond.L.Lock()
	b.closed = true
	b.cond.Broadcast()
	b.cond.L.Unlock()

	return nil
}
