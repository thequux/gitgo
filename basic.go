package gitgo

import "encoding/hex"

// A git file mode. Must be one of the following constants
type Filemode int

const (
	FilemodeDirectory  Filemode = 0040000
	FilemodeNormal              = 0100644
	FilemodeExecutable          = 0100755
	FilemodeSymlink             = 0120000
	FilemodeGitlink             = 0160000
)

type Oid [20]byte

func OidFromString(s string) (*Oid, error) {
	var ret Oid
	if len(s) != 40 {
		return nil, FormatError
	}
	if _, err := hex.Decode(ret[:], []byte(s)); err != nil {
		return nil, FormatError
	}
	return &ret, nil
}

func (oid *Oid) String() string {
	return hex.EncodeToString((*oid)[:])
}

func (oid *Oid) Equals(other *Oid) bool {
	for i := 0; i < 20; i++ {
		if (*oid)[i] != (*other)[i] {
			return false
		}
	}
	return true
}

////////////////////////////////////

type ObjectType int

const (
	TypeUnknown ObjectType = iota
	TypeBlob
	TypeTree
	TypeCommit
	TypeTag
)

var typeMap = map[string]ObjectType{
	"blob":   TypeBlob,
	"tree":   TypeTree,
	"commit": TypeCommit,
	"tag":    TypeTag,
}

func (o ObjectType) String() string {
	switch o {
	case TypeBlob:
		return "blob"
	case TypeTree:
		return "tree"
	case TypeCommit:
		return "commit"
	case TypeTag:
		return "tag"
	default:
		return "unkn"
	}
}
