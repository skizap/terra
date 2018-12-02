package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/stellarproject/terra/agent"
	"github.com/stellarproject/terra/version"
	"github.com/urfave/cli"
)

const (
	defaultGRPCPort    = 9005
	defaultClusterPort = 6946
)

func main() {
	app := cli.NewApp()
	app.Name = version.Name
	app.Version = version.BuildVersion()
	app.Author = "@stellarproject"
	app.Email = ""
	app.Usage = version.Description
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, D",
			Usage: "Enable debug logging",
		},
		cli.StringFlag{
			Name:  "node-id",
			Usage: "cluster node id",
			Value: getHostname(),
		},
		cli.StringFlag{
			Name:  "grpc-address, a",
			Usage: "address for grpc",
			Value: fmt.Sprintf("<computed>:9005"),
		},
		cli.StringFlag{
			Name:  "nic, n",
			Usage: "network interface to use for detecting IP (default: first non-local)",
			Value: "",
		},
		cli.StringFlag{
			Name:  "cluster-address",
			Usage: "cluster address",
			Value: fmt.Sprintf("<computed>:6946"),
		},
		cli.StringFlag{
			Name:  "advertise-address",
			Usage: "cluster advertise address",
			Value: fmt.Sprintf("<computed>:6946"),
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
	}
	app.Before = func(ctx *cli.Context) error {
		if ctx.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}

		// compute and assign addresses
		ip := getIP(ctx)
		if !ctx.IsSet("grpc-address") {
			ctx.Set("grpc-address", fmt.Sprintf("%s:%d", ip, defaultGRPCPort))
		}
		if !ctx.IsSet("cluster-address") {
			ctx.Set("cluster-address", fmt.Sprintf("%s:%d", ip, defaultClusterPort))
		}
		if !ctx.IsSet("advertise-address") {
			ctx.Set("advertise-address", fmt.Sprintf("%s:%d", ip, defaultClusterPort))
		}

		return nil
	}
	app.Action = agentAction

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
}

func agentAction(ctx *cli.Context) error {
	labels, err := getLabels(ctx.StringSlice("label"))
	if err != nil {
		return err
	}
	cfg := &agent.AgentConfig{
		NodeID:                ctx.String("node-id"),
		GRPCAddress:           ctx.String("grpc-address"),
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

	logrus.WithFields(logrus.Fields{
		"grpc":      cfg.GRPCAddress,
		"cluster":   cfg.ClusterAddress,
		"advertise": cfg.AdvertiseAddress,
	}).Info("starting terra agent")

	errCh := make(chan error)
	go func() {
		for err := range errCh {
			logrus.Error(err)
		}
	}()

	go func() {
		if err := a.Start(); err != nil {
			errCh <- err
		}
	}()

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	doneCh := make(chan bool, 1)
	go func() {
		for {
			select {
			case sig := <-signals:
				switch sig {
				case syscall.SIGTERM, syscall.SIGINT:
					logrus.Info("shutting down")
					if err := a.Stop(); err != nil {
						logrus.Error(err)
					}
					doneCh <- true
				default:
					logrus.Warnf("unhandled signal %s", sig)

				}

			}

		}

	}()

	<-doneCh
	return nil
}

func getIP(ctx *cli.Context) string {
	ip := "127.0.0.1"
	devName := ctx.String("nic")
	ifaces, err := net.Interfaces()
	if err != nil {
		logrus.Warnf("unable to detect network interfaces")
		return ip
	}

	for _, i := range ifaces {
		if devName == "" || i.Name == devName {
			a := getInterfaceIP(i)
			if a != "" {
				return a
			}
		}
	}

	logrus.Warnf("unable to find interface %s", devName)
	return ip
}

func getHostname() string {
	if h, _ := os.Hostname(); h != "" {
		return h
	}
	return "unknown"
}

func getInterfaceIP(iface net.Interface) string {
	addrs, err := iface.Addrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil || ip.IsLoopback() {
			continue
		}
		return ip.To4().String()
	}

	return ""
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
