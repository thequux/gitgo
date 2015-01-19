package main

import (
	"io"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/thequux/gitgo"
	"os"
)

var cmd_ListObjects = cli.Command{
	Name: "list_objects",
	ShortName: "lo",
	Action: func(context *cli.Context) {
		odb, err := gitgo.OpenRepository("")
		if err != nil {
			panic(err)
		}
		odb.Scan(func(oid *gitgo.Oid) error {
			fmt.Println(oid)
			return nil
		})
	},
}

var cmd_CatFile = cli.Command{
	Name: "cat-file",
	ShortName: "cat",
	Action: func(context *cli.Context) {
		odb, err := gitgo.OpenRepository("")
		if err != nil {
			panic(err)
		}
		oid, err := gitgo.OidFromString(context.Args().Get(0))
		if err != nil {
			panic(err)
		}
		obj, err := odb.Get(oid)
		if err != nil {
			panic(err)
		}
		io.Copy(os.Stdout, obj)
	},
}	
