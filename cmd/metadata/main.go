package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/carlosms/metadata-retrieval-playground/migration"
	v3 "github.com/carlosms/metadata-retrieval-playground/v3"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-cli.v0"
	"gopkg.in/src-d/go-log.v1"
)

// rewritten during the CI build step
var (
	version = "master"
	build   = "dev"
)

var app = cli.New("metadata", version, build, "GitHub metadata downloader")

func main() {
	app.AddCommand(&V3Command{})
	app.AddCommand(&MigrationCommand{})

	app.RunMain()
}

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

func setLogTransport(client *http.Client, logger log.Logger) {
	t := &logTransport{client.Transport, logger}
	client.Transport = t
}

type logTransport struct {
	T      http.RoundTripper
	Logger log.Logger
}

func (t *logTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	t0 := time.Now()

	resp, err := t.T.RoundTrip(r)
	if err != nil {
		return resp, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, err
	}

	t.Logger.With(
		log.Fields{"elapsed": time.Since(t0), "code": resp.StatusCode, "url": r.URL, "body": string(b)},
	).Debugf("HTTP response")

	resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))

	return resp, err
}
