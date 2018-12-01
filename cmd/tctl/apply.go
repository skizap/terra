package main

import (
	"encoding/json"
	"os"

	api "github.com/stellarproject/nebula/terra/v1"
	"github.com/urfave/cli"
)

var applyCommand = cli.Command{
	Name:      "apply",
	Usage:     "apply terra manifests",
	ArgsUsage: "[MANIFEST_LIST]",
	Flags:     []cli.Flag{},
	Action:    apply,
}

func apply(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	manifestListPath := ctx.Args().First()
	if _, err := os.Stat(manifestListPath); err != nil {
		return err
	}
	c, err := getClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	var manifestList *api.ManifestList
	f, err := os.Open(manifestListPath)
	if err != nil {
		return err
	}
	if err := json.NewDecoder(f).Decode(&manifestList); err != nil {
		return err
	}

	if err := c.Apply(manifestList.Manifests); err != nil {
		return err
	}

	return nil
}
