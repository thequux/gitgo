package gitgo

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strconv"
)

type TreeEntry struct {
	Name string
	Oid *Oid
	Mode Filemode
}
	
type Tree struct {
	paths map[string]TreeEntry
	repo *Repository
}

func (repo *Repository) ParseRawTree(obj io.Reader) (*Tree, error) {
	tree := Tree{
		repo: repo,
		paths: make(map[string]TreeEntry),
	}
	content, err := ioutil.ReadAll(obj)
	if err != nil {
		return nil, err
	}
	for len(content) != 0 {
		var entry TreeEntry
		modeEnd := bytes.Index(content, []byte{32})
		if modeEnd == -1 {
			return nil, FormatError
		}
		mode, err := strconv.ParseUint(string(content[:modeEnd]), 8, 16)
		if err != nil {
			fmt.Println(err)
			return nil, FormatError
		}
		entry.Mode = Filemode(mode)
		content = content[modeEnd+1:]

		nameEnd := bytes.Index(content, []byte{0})
		if nameEnd == -1 {
			return nil, FormatError
		}
		entry.Name = string(content[:nameEnd])
		content = content[nameEnd+1:]

		if len(content) < 20 {
			return nil, FormatError
		}
		entry.Oid = new(Oid)
		copy(entry.Oid[:], content)
		content = content[20:]
		
		tree.paths[entry.Name] = entry
	}
	return &tree, nil
}

type treeEntrySlice []TreeEntry
func (t treeEntrySlice) Less(i,j int) bool {
	name1 := t[i].Name
	if t[i].Mode == FilemodeDirectory {
		name1 = name1 + "/"
	}
	name2 := t[j].Name
	if t[j].Mode == FilemodeDirectory {
		name1 = name1 + "/"
	}

	return name1 < name2
}
func (t treeEntrySlice) Len() int {
	return len(t)
}

func (t treeEntrySlice) Swap(i,j int) {
	t[i], t[j] = t[j], t[i]
}

func (tree Tree) Dump(w io.Writer) error {
	entries := make(treeEntrySlice, 0, len(tree.paths))
	maxLength := 0
	for _, entry := range tree.paths {
		entries = append(entries, entry)
		if maxLength < len(entry.Name) {
			maxLength = len(entry.Name)
		}
	}
	sort.Sort(entries)
	for _, entry := range entries {
		if _, err := fmt.Fprintf(w, "%s %6o %s\n", entry.Oid, entry.Mode, entry.Name); err != nil {
			return err
		}
	}
	return nil
}
