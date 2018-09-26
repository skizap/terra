package install

import (
	"fmt"

	"github.com/stellarproject/terra/installer"
	"github.com/urfave/cli"
)

// Command is the install CLI command
var Command = cli.Command{
	Name:      "install",
	Usage:     "install assembly",
	ArgsUsage: "[IMAGE]",
	Action:    install,
}

func install(ctx *cli.Context) error {
	image := ctx.Args().First()
	if image == "" {
		cli.ShowSubcommandHelp(ctx)
		return fmt.Errorf("you must specify an image")
	}
	i := &installer.AssemblyInstaller{
		Image: image,
	}
	if err := i.Install(); err != nil {
		return err
	}
	return nil
}
