package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	api "github.com/stellarproject/nebula/terra/v1"
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
	fmt.Fprintf(w, "ID\tADDRESS\tLABELS\tSTATUS\n")
	for _, n := range nodes {
		labels := []string{}
		for k, v := range n.GetLabels() {
			labels = append(labels, fmt.Sprintf("%s=%s", k, v))
		}

		state := api.NodeStatus_Status_name[int32(n.GetStatus().Status)]
		status := fmt.Sprintf("%s", state)
		if desc := n.GetStatus().GetDescription(); desc != "" {
			status = fmt.Sprintf("%s (%s)", state, desc)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", n.GetID(), n.GetAddress(), strings.Join(labels, ","), status)
	}
	w.Flush()

	return nil
}
