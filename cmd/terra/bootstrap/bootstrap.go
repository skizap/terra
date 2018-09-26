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

var Command = cli.Command{
	Name:  "bootstrap",
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
		cli.StringFlag{
			Name:  "namespace, n",
			Usage: "registry namespace to use for base assemblies",
			Value: "stellarproject",
		},
		cli.StringFlag{
			Name:  "tag, t",
			Usage: "image tag to use for assemblies",
			Value: "latest",
		},
	},
	Action: bootstrap,
}

func bootstrap(ctx *cli.Context) error {
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
	if err := rf.Close(); err != nil {
		return err
	}

	if err := scp.Chmod(terraPath, 0755); err != nil {
		return err
	}

	// install assemblies
	namespace := ctx.String("namespace")
	tag := ctx.String("tag")
	assemblies := []string{
		"containerd",
		"buildkit",
		"cni-plugins",
	}

	for _, assembly := range assemblies {
		img := fmt.Sprintf("docker.io/%s/%s:%s", namespace, assembly, tag)
		cmd := terraPath + " install " + img
		out, err := sshConn.Exec(cmd)
		if err != nil {
			fmt.Fprintf(os.Stderr, string(out))
			return err
		}

		fmt.Fprintf(os.Stdout, string(out))
	}
	return nil
}
