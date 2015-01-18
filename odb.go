package gitgo

// TODO: Replace instances of Sprintf with a custom filepath formatting function

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type RawObject interface {
	Type() ObjectType

	// The length of the object in bytes; can also return -1 if
	// the length is unknown.  If the length happens to be known,
	// it can be written to disk without buffering the entire
	// object in memory.
	Length() int64
	io.Reader
}

// Odb is a low-level interface to the object database.  Chances are,
// you won't have a need for it.
type Odb interface {
	// Get an object from the ODB
	Get(oid *Oid) (RawObject, error)

	// Write an object to the ODB
	Put(obj RawObject) (*Oid, error)

	// Calls scanner once for each object in the database. If
	// scanner returns non-nil, returns immediately.  The *ONLY*
	// errors that this will report are errors from the scanner
	// function; it will silently succeed on internal errors.
	Scan(scanner func(oid *Oid) error) error
}

// An ODB consisting of loose files, e.g., a .git/objects directory.
// Content errors such as passing a non-directory or broken symlink
// are detected on creation. After opening, most errors are silently
// swallowed, unless there is no way to proceed without potentially
// losing data. (e.g., a plain file in place of a first-level
// directory is ignored for Get() (but still makes the Get() act as if
// no file was found), but Put() returns an error. Scan gives no
// fucks.
type LooseOdb struct {
	root string
}

var _ Odb = LooseOdb{}

func NewLooseOdb(path string) (Odb, error) {
	// TODO: Add basic validity checks on odb
	return LooseOdb{root: path}, nil
}

type looseOdbReader struct {
	size   int64
	otype  ObjectType
	reader io.Reader
}

func (l looseOdbReader) Type() ObjectType {
	return l.otype
}

func (l looseOdbReader) Length() int64 {
	return l.size
}

func (l looseOdbReader) Read(p []byte) (int, error) {
	return l.reader.Read(p)
}

func (odb LooseOdb) Get(oid *Oid) (RawObject, error) {
	if oid == nil {
		return nil, NilOidError
	}
	path := fmt.Sprintf("%s/%02x/%x", odb.root, (*oid)[0], (*oid)[1:])
	f, err := os.Open(path)
	if err == os.ErrNotExist {
		return nil, MissingObjectError
	}
	if err != nil {
		return nil, IoError
	}

	reader, err := zlib.NewReader(f)
	if err != nil {
		return nil, IoError
	}

	breader := bufio.NewReader(reader)
	header, err := breader.ReadString('\x00')
	if err != nil {
		return nil, IoError
	}
	header = header[:len(header)-1]
	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 {
		return nil, CorruptDbError
	}

	// If the type isn't recognized, this defaults to 0, or TypeUnknown
	otype := typeMap[headerParts[0]]
	size, err := strconv.ParseInt(headerParts[1], 10, 64)
	if err != nil {
		return nil, CorruptDbError
	}

	return looseOdbReader{
		size:   size,
		reader: breader,
		otype:  otype,
	}, nil
}

func (odb LooseOdb) Put(obj RawObject) (*Oid, error) {
	if obj.Length() < 0 {
		// Handle the indeterminate size case; after this,
		// Size() is >= 0
		buf := new(bytes.Buffer)
		size, err := io.Copy(buf, obj)
		if err != nil {
			return nil, err
		}
		obj = looseOdbReader{
			size:   size,
			otype:  obj.Type(),
			reader: buf,
		}
	}
	tmp, err := ioutil.TempFile(odb.root, "tempobj")
	defer func() {
		if tmp != nil {
			os.Remove(tmp.Name())
		}
	}()

	zlib_writer := zlib.NewWriter(tmp)
	hasher := sha1.New()
	w := io.MultiWriter(hasher, zlib_writer)

	fmt.Fprintf(w, "%s %d\x00", obj.Type(), obj.Length())
	// we ignore the error here, because hash won't throw errors,
	// and zlib saves them up
	written, _ := io.CopyN(w, obj, obj.Length())
	if written != obj.Length() {
		return nil, err
	}
	err = zlib_writer.Flush()
	if err != nil {
		return nil, err
	}
	var hash Oid
	copy(hash[:], hasher.Sum(nil))
	// TODO: Change file mode depending on Git config
	if err := os.MkdirAll(fmt.Sprintf("%s/%02x", odb.root, hash[0]), 0755); err != nil {
		return nil, err
	}

	if err := os.Rename(tmp.Name(), fmt.Sprintf("%s/%02x/%x", odb.root, hash[0], hash[1:])); err != nil {
		return nil, err
	}
	tmp = nil // don't try to delete a thing that doesn't exist.
	return &hash, nil
}

func (odb LooseOdb) Scan(scanner func(oid *Oid) error) error {
	tldir, err := os.Open(odb.root)
	if err != nil {
		return err
	}
	defer tldir.Close()
	if stat, err := tldir.Stat(); err != nil {
		return err
	} else if !stat.IsDir() {
		return nil
	}
	procDir := func(name string) error {
		var oid Oid
		if n, err := hex.Decode(oid[0:1], []byte(name)); n != 1 || err != nil {
			return nil
		}
		dName := filepath.Join(odb.root, name)
		dir, err := os.Open(dName)
		if err != nil {
			return nil
		} // We ignore all sorts of errors
		defer dir.Close()
		if stat, err := dir.Stat(); err != nil || !stat.IsDir() {
			return nil
		}
		fis, err := dir.Readdir(-1)
		if err != nil {
			return nil
		}
		for _, fi := range fis {
			if fi.IsDir() {
				continue
			}
			//fn := filepath.Join(dName, fi.Name())
			if len(fi.Name()) != 38 {
				continue
			}
			if n, err := hex.Decode(oid[1:], []byte(fi.Name())); err != nil || n != 19 {
				continue
			}
			if err := scanner(&oid); err != nil {
				return err
			}
		}
		return nil
	}
	tlfis, err := tldir.Readdir(-1)
	if err != nil {
		return nil
	}
	for _, tlfi := range tlfis {
		if !tlfi.IsDir() {
			continue
		}
		if len(tlfi.Name()) != 2 {
			continue
		}
		if err := procDir(tlfi.Name()); err != nil {
			return err
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////

type fullObj struct {
	Type    ObjectType
	Content []byte
}

type MemoryODB map[Oid]fullObj

func NewMemoryODB() Odb {
	return new(MemoryODB)
}

func (db *MemoryODB) Get(oid *Oid) (RawObject, error) {
	obj, ok := (*db)[*oid]
	if !ok {
		return nil, MissingObjectError
	}
	return looseOdbReader{
		size:   int64(len(obj.Content)),
		otype:  obj.Type,
		reader: bytes.NewReader(obj.Content),
	}, nil
}

func (db *MemoryODB) Put(obj RawObject) (*Oid, error) {
	content, err := ioutil.ReadAll(obj)
	if err != nil {
		return nil, err
	}
	hasher := sha1.New()
	fmt.Fprintf(hasher, "%s %d\x00", obj.Type(), len(content))
	var result Oid
	copy(result[:], hasher.Sum(content))
	(*db)[result] = fullObj{Type: obj.Type(), Content: content}
	return &result, nil
}

func (db *MemoryODB) Scan(scanner func(oid *Oid) error) error {
	for oid := range *db {
		if err := scanner(&oid); err != nil {
			return err
		}
	}
	return nil
}
