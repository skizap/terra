package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/sirupsen/logrus"
	"github.com/stellarproject/terra/cmd/terra/install"
	"github.com/stellarproject/terra/version"
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
	}
	app.Before = func(ctx *cli.Context) error {
		if ctx.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
