package hourtab

import (
	"bytes"

	ikea "github.com/ikkerens/ikeapack"
)

type Project struct {
	Path      string
	GitOrigin string

	Inactive   bool
	TotalTime  uint64
	ResetSince uint64
}

func (p *Project) Marshal() ([]byte, error) {
	var b bytes.Buffer

	if err := ikea.Pack(&b, p); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (p *Project) Unmarshal(b []byte) error {
	return ikea.Unpack(bytes.NewReader(b), p)
}
