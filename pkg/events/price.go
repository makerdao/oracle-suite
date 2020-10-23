package events

import (
	"encoding/json"

	"github.com/makerdao/gofer/internal/oracle"
)

const PriceEvent = "price"

type Price struct {
	Price *oracle.Price   `json:"price"`
	Trace json.RawMessage `json:"trace"`
}

func (p *Price) PayloadMarshall() ([]byte, error) {
	p.Trace = []byte("null") // remove for now

	// TODO: use binary format and base64 to reduce payload size
	return json.Marshal(p)
}

func (p *Price) PayloadUnmarshall(b []byte) error {
	return json.Unmarshal(b, p)
}
