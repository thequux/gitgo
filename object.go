package gitgo

import "io"

type GitObject interface {
	Dump(w io.Writer) error
}

func (repo *Repository) ParseObject(obj RawObject) (GitObject, error) {
	switch obj.Type() {
	case TypeTree:
		return repo.ParseRawTree(obj)
	default:
		return nil, NotImplementedError
	}
}
