package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli"
)

var clusterCommand = cli.Command{
	Name:  "cluster",
	Usage: "cluster operations",
	Subcommands: []cli.Command{
		nodesCommand,
	},
}

var nodesCommand = cli.Command{
	Name:   "nodes",
	Usage:  "list terra nodes",
	Flags:  []cli.Flag{},
	Action: nodes,
}

func nodes(ctx *cli.Context) error {
	c, err := getClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	nodes, err := c.Nodes()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	fmt.Fprintf(w, "ID\tADDRESS\tLABELS\n")
	for _, n := range nodes {
		labels := []string{}
		for k, v := range n.Labels {
			labels = append(labels, fmt.Sprintf("%s=%s", k, v))
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", n.ID, n.Address, strings.Join(labels, ","))
	}
	w.Flush()

	return nil
}
