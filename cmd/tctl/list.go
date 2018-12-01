package main

import (
	"html/template"
	"os"

	"github.com/urfave/cli"
)

const (
	listTemplate = `{{ range .Manifests }}- NodeID: {{ .NodeID }} {{ if .Labels }}
  Labels: {{ .Labels }}{{ end }}
  Assemblies:
{{ range .Assemblies }}    - Image: {{ .Image }}
{{ end }}
{{ end }}`
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
