package main

import (
	"encoding/json"
	"html/template"
	"os"

	api "github.com/stellarproject/nebula/terra/v1"
	"github.com/urfave/cli"
)

const (
	listTemplate = `{{ range .Manifests }}- NodeID: {{ .NodeID }}{{ if .Labels }}
  Labels: {{ range $k, $v := .Labels }}
    - {{ $k }}={{ $v }}{{ end }}{{ end }}
  Assemblies:
{{ range .Assemblies }}    - Image: {{ .Image }}
{{ end }}
{{ end }}`
)

var manifestCommand = cli.Command{
	Name:  "manifest",
	Usage: "manifest operations",
	Subcommands: []cli.Command{
		listCommand,
		applyCommand,
	},
}

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

	manifestList, err := c.List()
	if err != nil {
		return err
	}

	if manifestList == nil {
		return nil
	}

	t := template.New("list")
	tmpl, err := t.Parse(listTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(os.Stdout, manifestList); err != nil {
		return err
	}

	return nil
}

var applyCommand = cli.Command{
	Name:      "apply",
	Usage:     "apply terra manifests",
	ArgsUsage: "[MANIFEST_LIST]",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force",
			Usage: "force apply manifest list",
		},
	},
	Action: apply,
}

func apply(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	manifestListPath := ctx.Args().First()
	force := ctx.Bool("force")
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

	if err := c.Apply(manifestList.Manifests, force); err != nil {
		return err
	}

	return nil
}
