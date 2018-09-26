package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/stellarproject/terra/cmd/terra/bootstrap"
	"github.com/stellarproject/terra/cmd/terra/install"
	"github.com/stellarproject/terra/version"
	"github.com/urfave/cli"
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
	}
	app.Commands = []cli.Command{
		install.Command,
		bootstrap.Command,
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
