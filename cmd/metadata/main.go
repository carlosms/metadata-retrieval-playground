package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"database/sql"

	"github.com/carlosms/metadata-retrieval-playground/cmd/metadata/subcmd"
	"github.com/carlosms/metadata-retrieval-playground/migration"
	v3 "github.com/carlosms/metadata-retrieval-playground/v3"
	v4 "github.com/carlosms/metadata-retrieval-playground/v4"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
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
	app.AddCommand(&subcmd.LimitsCommand{})
	app.AddCommand(&V3Command{})
	app.AddCommand(&V4Command{})
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

type V4Command struct {
	cli.Command `name:"v4" short-description:"" long-description:""`

	LogHTTP bool `long:"log-http" description:"log http requests (debug level)"`

	DB    string `long:"db" description:"PostgreSQL URL connection string, e.g. postgres://user:password@127.0.0.1:5432/ghsync?sslmode=disable"`
	Token string `long:"token" short:"t" env:"SOURCED_GITHUB_TOKEN" description:"GitHub personal access token" required:"true"`

	Owner string `long:"owner"  required:"true"`
	Name  string `long:"name"  required:"true"`
}

func (c *V4Command) Execute(args []string) error {
	client := oauth2.NewClient(context.TODO(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.Token},
	))

	if c.LogHTTP {
		logger := log.New(log.Fields{"owner": c.Owner, "repo": c.Name})
		setLogTransport(client, logger)
	}

	var downloader *v4.GitHubDownloader
	if c.DB == "" {
		log.Infof("using stdout to save the data")
		var err error
		downloader, err = v4.NewStdoutDownloader(client)
		if err != nil {
			return err
		}
	} else {
		db, err := sql.Open("postgres", c.DB)
		if err != nil {
			return err
		}

		defer func() {
			if err != nil {
				db.Close()
				db = nil
			}
		}()

		if err = db.Ping(); err != nil {
			return err
		}

		if err = c.dbMigrate(); err != nil && err != migrate.ErrNoChange {
			return err
		}

		downloader, err = v4.NewDBDownloader(client, db)
	}

	return downloader.DownloadRepository(c.Owner, c.Name, "v0")
}

func (c *V4Command) dbMigrate() error {
	m, err := migrate.New(
		"file://v4/db/migrations",
		c.DB)
	if err != nil {
		return err
	}
	return m.Up()
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
