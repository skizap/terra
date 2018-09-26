package install

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/pkg/errors"
	"github.com/stellarproject/terra/installer"
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
	out, err := i.Install()
	if err != nil {
		return errors.Wrap(err, string(out))
	}
	fmt.Fprintf(os.Stdout, string(out))
	return nil
}
