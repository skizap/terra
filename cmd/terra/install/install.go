package install

import (
	"io"
	"os"

	"github.com/codegangsta/cli"
	"github.com/pkg/sftp"
	"github.com/stellarproject/terra/utils"
)

var Command = cli.Command{
	Name:  "install",
	Usage: "install system components",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "user, u",
			Usage: "user to use for installation",
			Value: "root",
		},
		cli.StringFlag{
			Name:  "host",
			Usage: "host to target for installation",
			Value: "",
		},
	},
	Before: func(ctx *cli.Context) error {
		host := ctx.String("host")
		if host == "" {
			return nil
		}
		user := ctx.String("user")
		// copy binary to remote
		sshConn, err := utils.NewSSHClient(user, host)
		if err != nil {
			return err
		}

		scp, err := sftp.NewClient(sshConn.Client())
		if err != nil {
			return err
		}
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		ef, err := os.Open(exe)
		if err != nil {
			return err
		}

		rf, err := scp.OpenFile("/usr/local/bin/terra", os.O_WRONLY)
		if err != nil {
			return err
		}

		if _, err := io.Copy(rf, ef); err != nil {
			return err
		}

		return nil
	},
	Subcommands: []cli.Command{
		installContainerdCommand,
	},
}
