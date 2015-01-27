package gitgo

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Repository struct {
	root string
	odb  []Odb
}

func exists(path ...string) bool {
	_, err := os.Stat(filepath.Join(path...))
	return err == nil
}

func is_dir(path ...string) bool {
	stat, err := os.Stat(filepath.Join(path...))
	return err == nil && stat.IsDir()
}

func git_dir_is_valid(path string) bool {
	return (is_dir(path) &&
		is_dir(path, "refs") &&
		is_dir(path, "objects") &&
		exists(path, "HEAD"))
}

// Find the git directory that would be used if the reference Git
// implementation were run from `path`. If path is empty, computes the
// default git directory (including considering the GIT_DIR env var)
func Discover(path string) (string, error) {
	if path == "" {
		// Check GIT_DIR
		var err error
		git_dir := os.Getenv("GIT_DIR")
		if git_dir != "" {
			if !exists(git_dir) {
				return "", DbNotFoundError
			}
			path, err = filepath.Abs(git_dir)
			if err != nil {
				return "", err
			}
		} else {
			path, err = os.Getwd()
			if err != nil {
				return "", err
			}
		}
	} else {
		var err error
		path, err = filepath.Abs(path)
		if err != nil {
			return "", err
		}
	}
	// Start walking upwards...
	for ; path != "/"; path = filepath.Dir(path) {
		if exists(path, ".git") {
			path = filepath.Join(path, ".git")
			goto resolve_path
		}
		if git_dir_is_valid(path) {
			goto resolve_path
		}
	}
	return "", DbNotFoundError
resolve_path:
	stat, err := os.Stat(path)
	if err != nil {
		return "", DbNotFoundError
	}
	if stat.Mode().IsRegular() {
		prefix := `gitdir:`
		gitfile_content, err := ioutil.ReadFile(path)
		if err != nil {
			return "", err
		}
		if len(gitfile_content) < len(prefix) {
			return "", CorruptDbError
		}
		if string(gitfile_content[:len(prefix)]) != prefix {
			return "", CorruptDbError
		}
		path = filepath.Join(filepath.Dir(path),
			strings.Trim(string(gitfile_content[len(prefix):]),
				" \t\n"))
		goto resolve_path
	}
	if !stat.IsDir() {
		return "", FormatError
	}
	return filepath.Abs(path)
}

func OpenRepository(path string) (*Repository, error) {
	// TODO: implement me
	if !git_dir_is_valid(path) {
		return nil, DbNotFoundError
	}

	odbs := []Odb{}
	odb, err := NewLooseOdb(filepath.Join(path, "objects"))
	odbs = append(odbs, odb)
	if err != nil {
		return nil, err
	}
	return &Repository{
		root: path,
		odb: odbs,
	}, nil
}

// Meta-implementation of ODB...
var _ Odb = &Repository{}

func (repo *Repository) Get(oid *Oid) (RawObject, error) {
	for _, odb := range repo.odb {
		if obj, err := odb.Get(oid); err == nil {
			return obj, err
		}
	}
	return nil, MissingObjectError
}
	
func (repo *Repository) Put(obj RawObject) (*Oid, error) {
	return repo.odb[0].Put(obj)
}

func (repo *Repository) Scan(scanner func(*Oid) error) error {
	for _, odb := range repo.odb {
		if err := odb.Scan(scanner); err != nil {
			return err
		}
	}
	return nil
}
