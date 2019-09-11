package main

import (
	"github.com/carlosms/metadata-retrieval-playground/cmd/metadata/subcmd"
	"gopkg.in/src-d/go-cli.v0"
)

// rewritten during the CI build step
var (
	version = "master"
	build   = "dev"
)

var app = cli.New("metadata", version, build, "GitHub metadata downloader")

func main() {
	app.AddCommand(&subcmd.LimitsCommand{})
	app.AddCommand(&subcmd.V3Command{})
	app.AddCommand(&subcmd.V4Command{})
	app.AddCommand(&subcmd.MigrationCommand{})

	app.RunMain()
}
