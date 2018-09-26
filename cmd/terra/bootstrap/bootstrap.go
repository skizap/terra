package bootstrap

import (
	"github.com/codegangsta/cli"
)

var Command = cli.Command{
	Name:  "bootstrap",
	Usage: "initialize terra",
	Subcommands: []cli.Command{
		initCommand,
	},
}
