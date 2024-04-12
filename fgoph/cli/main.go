package main

import (
	"log"
	"os"
	"sort"

	"github.com/urfave/cli"
)

/*
This is the CLI to execute the main management modules of firegopher
*/

//@TODO Add a dependency checker

func main() {

	app := cli.NewApp()

	app.Name = "firegopher"
	app.Usage = "deploy apps with crack, not containers"
	app.Author = "giesf"
	app.Version = "v0.0.0"

	app.Flags = []cli.Flag{}
	app.UseShortOptionHandling = true

	app.Commands = []cli.Command{
		runCommand,
		bootstrapCmd,
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	args := preprocessArgs(os.Args)

	appErr := app.Run(args)
	if appErr != nil {
		log.Fatal(appErr)
	}
}

func preprocessArgs(args []string) []string {
	for i, arg := range args {
		if arg == "--" {
			args[i] = "__DASH_DASH__" // Replace with a placeholder
		}
	}
	return args
}
