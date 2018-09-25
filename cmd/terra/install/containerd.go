package install

import (
	"github.com/codegangsta/cli"
	"github.com/stellarproject/terra/installer/containerd"
)

var installContainerdCommand = cli.Command{
	Name:  "containerd",
	Usage: "install and upgrade containerd",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "image",
			Usage: "containerd image name",
			Value: "docker.io/stellarproject/containerd:latest",
		},
		cli.StringFlag{
			Name:  "dest",
			Usage: "installation dest path",
			Value: "/",
		},
	},
	Action: installContainerd,
}

func installContainerd(ctx *cli.Context) error {
	i := &containerd.ContainerdInstaller{
		Image:    ctx.String("image"),
		DestPath: ctx.String("dest"),
	}
	if err := i.Install(); err != nil {
		return err
	}
	return nil
}
