package main

import (
	"github.com/codegangsta/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Version =  "0.0.1"
	app.Commands = []cli.Command{
		cmd_CatFile,
		cmd_ListObjects,
	}
	app.Run(os.Args)
}
