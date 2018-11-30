package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/stellarproject/terra/client"
	"github.com/stellarproject/terra/version"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = version.Name + " cli"
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
			Name:  "addr, a",
			Usage: "address for grpc",
			Value: "127.0.0.1:9005",
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
	}
	app.Commands = []cli.Command{
		listCommand,
		applyCommand,
	}
	app.Before = func(ctx *cli.Context) error {
		if ctx.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
}

func getClient(ctx *cli.Context) (*client.Client, error) {
	addr := ctx.GlobalString("addr")
	return client.NewClient(addr)
}
