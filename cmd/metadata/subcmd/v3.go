package subcmd

import (
	"context"

	v3 "github.com/carlosms/metadata-retrieval-playground/v3"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-cli.v0"
	"gopkg.in/src-d/go-log.v1"
)

type V3Command struct {
	cli.Command `name:"v3" short-description:"" long-description:""`

	LogHTTP bool `long:"log-http" description:"log http requests (debug level)"`

	//DB    string `long:"db" description:"PostgreSQL URL connection string" required:"true"`
	Token string `long:"token" short:"t" env:"SOURCED_GITHUB_TOKEN" description:"GitHub personal access token" required:"true"`

	Owner string `long:"owner"  required:"true"`
	Name  string `long:"name"  required:"true"`
}

func (c *V3Command) Execute(args []string) error {
	client := oauth2.NewClient(context.TODO(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.Token},
	))

	if c.LogHTTP {
		logger := log.New(log.Fields{"owner": c.Owner, "repo": c.Name})
		setLogTransport(client, logger)
	}

	downloader, err := v3.NewStdoutDownloader(client)
	if err != nil {
		return err
	}

	return downloader.DownloadRepository(c.Owner, c.Name, "v0")
}
