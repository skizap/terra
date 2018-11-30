package main

import (
	"fmt"

	"github.com/urfave/cli"
)

var listCommand = cli.Command{
	Name:   "list",
	Usage:  "list terra assemblies",
	Flags:  []cli.Flag{},
	Action: list,
}

func list(ctx *cli.Context) error {
	c, err := getClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	assemblies, err := c.List()
	if err != nil {
		return err
	}

	fmt.Println(assemblies)

	return nil
}
