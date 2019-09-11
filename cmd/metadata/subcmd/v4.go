package subcmd

import (
	"context"
	"database/sql"
	"time"

	v4 "github.com/carlosms/metadata-retrieval-playground/v4"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-cli.v0"
	"gopkg.in/src-d/go-log.v1"
)

type V4Command struct {
	cli.Command `name:"v4" short-description:"" long-description:""`

	LogHTTP bool `long:"log-http" description:"log http requests (debug level)"`

	DB      string `long:"db" description:"PostgreSQL URL connection string, e.g. postgres://user:password@127.0.0.1:5432/ghsync?sslmode=disable"`
	Token   string `long:"token" short:"t" env:"SOURCED_GITHUB_TOKEN" description:"GitHub personal access token" required:"true"`
	Version string `long:"version" description:"Version tag in the DB"`

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

	version := c.Version
	if version == "" {
		version = time.Now().Format("2006-01-02 15:04:05")
	}
	err := downloader.DownloadRepository(c.Owner, c.Name, version)
	if err != nil {
		return err
	}

	return downloader.SetCurrent(version)
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
