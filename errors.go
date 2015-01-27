package gitgo

import (
	"errors"
)

type GitError string

func (e GitError) Error() string {
	return string(e)
}

var (
	FormatError         = errors.New("Format error")
	NilOidError         = errors.New("Nil OID")
	IoError             = errors.New("I/O error")
	CorruptDbError      = errors.New("Corrupt DB")
	MissingObjectError  = errors.New("Missing object")
	DbNotFoundError     = errors.New("No DB found")
	NotImplementedError = errors.New("Not implemented")
)
