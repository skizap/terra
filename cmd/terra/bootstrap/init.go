package bootstrap

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/pkg/sftp"
	"github.com/stellarproject/terra/utils"
)

const (
	terraPath = "/usr/local/bin/terra"
)

var initCommand = cli.Command{
	Name:  "init",
	Usage: "initialize terra on remote host",
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
	Action: initAction,
}

func initAction(ctx *cli.Context) error {
	host := ctx.String("host")
	user := ctx.String("user")
	if host == "" || user == "" {
		cli.ShowSubcommandHelp(ctx)
		return fmt.Errorf("host and user must be specified")
	}
	// check for port and add default if missing
	if !strings.Contains(host, ":") {
		host += ":22"
	}
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

	rf, err := scp.OpenFile(terraPath, os.O_WRONLY)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		f, err := scp.Create(terraPath)
		if err != nil {
			return err
		}

		rf = f
	}

	if _, err := io.Copy(rf, ef); err != nil {
		return err
	}

	if err := scp.Chmod(terraPath, 0755); err != nil {
		return err
	}

	return nil
}
