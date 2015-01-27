package gitgo

import (
	"io"
	"io/ioutil"
)

// TODO: Allow streaming read to handle large objects better

type Blob struct {
	content []byte
}

func ParseRawBlob(obj RawObject) (*Blob, error) {
	content, err := ioutil.ReadAll(obj)
	if err != nil {
		return nil, err
	} else {
		return &Blob{content}, nil
	}
}

func (b *Blob) Content() []byte {
	return b.content
}

func (b *Blob) Dump(w io.Writer) error {
	_, err := w.Write(b.content)
	return err
}
