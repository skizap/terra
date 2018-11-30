package main

import (
	"fmt"
	"time"

	api "github.com/stellarproject/nebula/terra/v1"
	"github.com/urfave/cli"
)

var applyCommand = cli.Command{
	Name:   "apply",
	Usage:  "apply terra manifests",
	Flags:  []cli.Flag{},
	Action: apply,
}

func apply(ctx *cli.Context) error {
	c, err := getClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	manifests := []*api.Manifest{
		{
			NodeID: "test",
			Assemblies: []*api.Assembly{
				{
					Image: fmt.Sprintf("r.local.io/nodes/test:%s", time.Now().String()),
				},
			},
		},
	}
	if err := c.Apply(manifests); err != nil {
		return err
	}

	return nil
}
