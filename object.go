package gitgo

import (
	"fmt"
	"io"
)

var _ = fmt.Print

type GitObject interface {
	Dump(w io.Writer) error
}

func (repo *Repository) ParseObject(obj RawObject) (GitObject, error) {
	switch obj.Type() {
	case TypeTree:
		return repo.ParseRawTree(obj)
	case TypeBlob:
		return ParseRawBlob(obj)
	default:
		//fmt.Printf("Unimplemented type %s", obj.Type())
		return nil, NotImplementedError
	}
}
