package query

import "encoding/json"

type JSON struct {
	data interface{}
}

func Parse(j []byte) (*JSON, error) {
	r := &JSON{}
	err := json.Unmarshal(j, r.data)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (j *JSON) get(path string)  {

}
