package gitgo

import (
	"bufio"
	"io"
	"time"
	"strings"
	"regexp"
	"strconv"
	"io/ioutil"
	"fmt"
)

type Signature struct {
	Name  string
	Email string
	Date  time.Time
}

func (s Signature) String() string {
	return fmt.Sprintf("%s <%s> %s", s.Name, s.Email, s.Date.Format(time.RFC1123))
}

type Commit struct {
	Tree      *Oid
	Parents   []*Oid
	Author    Signature
	Committer Signature
	Message   string
}

var safeStringRE = strings.Replace(strings.Replace(`[^ .,:;<>"'\0\n]|[^ .,:;<>"'\0][^\0\n<>]*[^ .,:;<>"'\0\n]`, "\\0", "\000", -1), `\n`, "\n", -1)
// groups: name email date_seconds date_tz
var signatureRE = regexp.MustCompile(`^(` + safeStringRE +`) <(` + safeStringRE + `)> (\d+) ([+-](?:0[0-9]|1[012])(?:[0-5][0-9]))$`)

func parseSignature(str string) (Signature, error) {
	results := signatureRE.FindStringSubmatch(str)
	
	if results == nil {
		return Signature{}, FormatError
	}
	var ret Signature
	ret.Name = results[1]
	ret.Email = results[2]

	date_s, _ := strconv.ParseInt(results[3], 10, 32)
	tz_hr, _ := strconv.ParseInt(results[4][1:3], 10, 8)
	tz_min, _ := strconv.ParseInt(results[4][3:5], 10, 8)
	var tz_sign int64
	if results[4][0] == '-' {
		tz_sign = -1
	} else {
		tz_sign = 1
	}

	tz_s := tz_sign * 60 * (tz_hr * 60 + tz_min)
	ret.Date = time.Unix(date_s, 0).In(time.FixedZone(results[4], int(tz_s)))
	return ret, nil
}

func (repo *Repository) ParseRawCommit(reader io.Reader) (*Commit, error) {
	r := bufio.NewReader(reader)

	ret := &Commit{}

	have_author := false
	have_commiter := false
	for {
		line, err := r.ReadString(0x0a)
		if err != nil {
			return nil, err
		}
		line = strings.Trim(line, "\n")
		if len(line) == 0 {
			// End of header
			break
		}
		
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			return nil, FormatError
		}
		// Strictly speaking, according to the grammar at
		// http://git.rsbx.net/Documents/Git_Data_Formats.txt
		// there is a specific order to the lines. There may
		// be other implementations that ignore it, and so
		// long as lines that shouldn't be repeated aren't,
		// there's no problem with reordering the lines,
		// particularly because this is not security-critical.
		// We *do* generate the lines in the right order
		// though.
		switch parts[0] {
		case "tree":
			if ret.Tree != nil {
				// only one tree is allowed
				return nil, FormatError
			}
			ret.Tree, err = OidFromString(parts[1])
		case "parent":
			oid, err := OidFromString(parts[1])
			if err != nil {
				return nil, FormatError
			}
			ret.Parents = append(ret.Parents, oid)
		case "author":
			if have_author {
				return nil, FormatError
			}
			ret.Author, err = parseSignature(parts[1])
			if err != nil {
				return nil, FormatError
			}
		case "committer":
			if have_commiter {
				return nil, FormatError
			}
			ret.Committer, err = parseSignature(parts[1])
			if err != nil {
				return nil, FormatError
			}
		default:
			// Unknown tag
			return nil, FormatError
		}
	}

	msg, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	ret.Message = string(msg)
	return ret, nil
}

func (c *Commit) Dump(w io.Writer) error {
	fmt.Fprintln(w, "tree ", c.Tree)
	for _, parent := range c.Parents {
		fmt.Fprintln(w, "parent ", parent)
	}
	fmt.Fprintln(w, "author ", c.Author)
	fmt.Fprintln(w, "committer ", c.Committer)

	fmt.Fprint(w, "\n", c.Message)
	return nil
}
