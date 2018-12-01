package main

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/stellarproject/terra/agent"
	"github.com/urfave/cli"
)

var agentCommand = cli.Command{
	Name:  "agent",
	Usage: "run terra agent",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "node-id",
			Usage: "cluster node id",
			Value: getHostname(),
		},
		cli.StringFlag{
			Name:  "addr, a",
			Usage: "address for grpc",
			Value: "127.0.0.1:9005",
		},
		cli.StringFlag{
			Name:  "cluster-address",
			Usage: "cluster address",
			Value: "127.0.0.1:6946",
		},
		cli.StringFlag{
			Name:  "advertise-address",
			Usage: "cluster advertise address",
			Value: "127.0.0.1:6946",
		},
		cli.StringFlag{
			Name:  "connection-type",
			Usage: "connection type (lan, wan, local)",
			Value: "local",
		},
		cli.StringSliceFlag{
			Name:  "peer",
			Usage: "cluster peer(s) to join",
			Value: &cli.StringSlice{},
		},
		cli.StringSliceFlag{
			Name:  "label",
			Usage: "node label(s)",
			Value: &cli.StringSlice{},
		},
		cli.StringFlag{
			Name:  "tls-cert",
			Usage: "tls certificate",
			Value: "",
		},
		cli.StringFlag{
			Name:  "tls-key",
			Usage: "tls key",
			Value: "",
		},
		cli.BoolFlag{
			Name:  "tls-insecure-skip-verify",
			Usage: "skip tls verification",
		},
		cli.StringFlag{
			Name:  "data-dir",
			Usage: "terra agent data directory",
			Value: "/var/lib/terra",
		},
	},
	Action: agentAction,
}

func agentAction(ctx *cli.Context) error {
	labels, err := getLabels(ctx.StringSlice("label"))
	if err != nil {
		return err
	}
	cfg := &agent.AgentConfig{
		NodeID:                ctx.String("node-id"),
		GRPCAddress:           ctx.String("addr"),
		ClusterAddress:        ctx.String("cluster-address"),
		AdvertiseAddress:      ctx.String("advertise-address"),
		ConnectionType:        ctx.String("connection-type"),
		DataDir:               ctx.String("data-dir"),
		Peers:                 ctx.StringSlice("peer"),
		Labels:                labels,
		TLSServerCertificate:  ctx.String("tls-cert"),
		TLSServerKey:          ctx.String("tls-key"),
		TLSInsecureSkipVerify: ctx.Bool("tls-insecure-skip-verify"),
	}
	a, err := agent.NewAgent(cfg)
	if err != nil {
		return err
	}

	logrus.Infof("starting terra agent on %s", cfg.GRPCAddress)

	return a.Start()
}

func getHostname() string {
	if h, _ := os.Hostname(); h != "" {
		return h
	}
	return "unknown"
}

func getLabels(lbls []string) (map[string]string, error) {
	labels := map[string]string{}
	for _, kv := range lbls {
		parts := strings.SplitN(kv, "=", 2)
		key := parts[0]

		if len(parts) == 0 {
			labels[key] = ""
			continue
		}

		labels[key] = strings.Join(parts[1:], "")
	}

	return labels, nil
}
