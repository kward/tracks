package commands

import "github.com/urfave/cli"

var commands []cli.Command

// Commands returns the supported cli commands.
func Commands() []cli.Command { return commands }
