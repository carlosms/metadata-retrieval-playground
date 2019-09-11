package subcmd

import (
	"context"

	"github.com/carlosms/metadata-retrieval-playground/migration"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-cli.v0"
	"gopkg.in/src-d/go-log.v1"
)

type MigrationCommand struct {
	cli.Command `name:"migration" short-description:"" long-description:""`

	LogHTTP bool `long:"log-http" description:"log http requests (debug level)"`

	//DB    string `long:"db" description:"PostgreSQL URL connection string" required:"true"`
	Token string `long:"token" short:"t" env:"SOURCED_GITHUB_TOKEN" description:"GitHub personal access token" required:"true"`

	Owner string `long:"owner"  required:"true"`
	Name  string `long:"name"  required:"true"`
}

func (c *MigrationCommand) Execute(args []string) error {
	client := oauth2.NewClient(context.TODO(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.Token},
	))

	if c.LogHTTP {
		logger := log.New(log.Fields{"owner": c.Owner, "repo": c.Name})
		setLogTransport(client, logger)
	}

	downloader, err := migration.NewMigrationDownloader(client)
	if err != nil {
		return err
	}

	return downloader.DownloadRepository(c.Owner, c.Name, "v0")
}
