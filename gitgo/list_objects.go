package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/thequux/gitgo"
	"io"
	"os"
)

func init() {
	registerCommand(cli.Command{
		Name:      "list_objects",
		ShortName: "lo",
		Action: func(context *cli.Context) {
			path, err := gitgo.Discover("")
			if err != nil {
				panic(err)
			}
			repo, err := gitgo.OpenRepository(path)
			if err != nil {
				panic(err)
			}
			repo.Scan(func(oid *gitgo.Oid) error {
				fmt.Println(oid)
				return nil
			})
		},
	})

	registerCommand(cli.Command{
		Name:      "cat-file",
		ShortName: "cat",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name: "r",
				Usage: "Dump raw object content",
			},
		},
		Action: func(context *cli.Context) {
			path, err := gitgo.Discover("")
			repo, err := gitgo.OpenRepository(path)
			if err != nil {
				panic(err)
			}
			oid, err := gitgo.OidFromString(context.Args().Get(0))
			if err != nil {
				panic(err)
			}
			obj, err := repo.Get(oid)
			if err != nil {
				panic(err)
			}
			cobj, err := repo.ParseObject(obj)
			if err == gitgo.NotImplementedError {
				fmt.Println("Not implemented")
				obj, err = repo.Get(oid)
				io.Copy(os.Stdout, obj)
			} else if err != nil {
				fmt.Printf("Error: %s\n", err)
			} else {
				cobj.Dump(os.Stdout)
			}
		},
	})
}
