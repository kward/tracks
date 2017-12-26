package main

import (
	"os"

	"github.com/kward/golib/os/sysexits"
	"github.com/kward/tracks/commands"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Commands = commands.Commands()
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "dry_run,n",
			Usage: "do a dry run",
		},
	}
	app.Name = "tracks - A tool for integrating Waves Tracks and Avid Venue"
	app.Usage = ""
	app.Version = ""
	app.Run(os.Args)
	os.Exit(sysexits.OK.Int())
}
