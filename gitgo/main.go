package main

import (
	"github.com/codegangsta/cli"
	"os"
)

var commands = []cli.Command{}
func registerCommand(cmd cli.Command) {
	commands = append(commands, cmd)
}

func main() {
	app := cli.NewApp()
	app.Version =  "0.0.1"
	app.Commands = commands
	app.Run(os.Args)
}
