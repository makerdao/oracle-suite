package graph

import "encoding/json"

func (p OriginTick) MarshalJSON() ([]byte, error) {
	var errStr string
	if p.Error != nil {
		errStr = p.Error.Error()
	}

	return json.Marshal(struct {
		Tick
		Origin string
		Error  string
	}{
		Tick:   p.Tick,
		Origin: p.Origin,
		Error:  errStr,
	})
}

func (p IndirectTick) MarshalJSON() ([]byte, error) {
	var errStr string
	if p.Error != nil {
		errStr = p.Error.Error()
	}

	return json.Marshal(struct {
		Tick
		OriginTicks   []OriginTick
		IndirectTicks []IndirectTick
		Error         string
	}{
		Tick:          p.Tick,
		OriginTicks:   p.OriginTicks,
		IndirectTicks: p.IndirectTicks,
		Error:         errStr,
	})
}
