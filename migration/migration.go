package migration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/carlosms/metadata-retrieval-playground"
	"github.com/carlosms/metadata-retrieval-playground/internal/client"
	"github.com/google/go-github/github"
	"github.com/mholt/archiver"
	"gopkg.in/src-d/go-log.v1"
)

type GitHubMigrationDownloader struct {
	//storer

	client *github.Client
	token  string
}

var _ metadata.MetadataDownloader = GitHubMigrationDownloader{}

func NewMigrationDownloader(httpClient *http.Client) (*GitHubMigrationDownloader, error) {
	c, err := client.NewClient(httpClient)
	if err != nil {
		return nil, err
	}

	return &GitHubMigrationDownloader{
		client: c,
	}, nil
}

func (d GitHubMigrationDownloader) DownloadRepository(owner string, name string, version string) error {
	logger := log.New(log.Fields{"owner": owner, "repo": name})

	t0 := time.Now()

	opt := github.MigrationOptions{
		LockRepositories:   false,
		ExcludeAttachments: true,
	}
	migration, _, err := d.client.Migrations.StartMigration(context.TODO(), owner, []string{name}, &opt)
	if err != nil {
		return err
	}

	logger = logger.With(log.Fields{"migration-id": migration.GetID()})
	logger.With(log.Fields{"state": migration.GetState()}).Infof("migration started")

	// pending, which means the migration hasn't started yet.
	// exporting, which means the migration is in progress.
	// exported, which means the migration finished successfully.
	// failed, which means the migration failed.

	for migration.GetState() != "exported" {
		logger.With(log.Fields{"state": migration.GetState()}).Infof("waiting for migration to be ready")
		time.Sleep(time.Second)
		migration, _, err = d.client.Migrations.MigrationStatus(context.TODO(), owner, migration.GetID())

		if migration.GetState() == "failed" {
			return fmt.Errorf("migration %v for organization %v returned state 'failed'", migration.GetID(), owner)
		}
	}

	url, err := d.client.Migrations.MigrationArchiveURL(context.TODO(), owner, migration.GetID())
	if err != nil {
		return err
	}

	elapsed := time.Since(t0)
	logger.With(log.Fields{"elapsed": elapsed, "state": migration.GetState(), "url": url}).Infof("migration ready to download")

	t1 := time.Now()

	path := filepath.Join("downloads", fmt.Sprintf("%v-%v", owner, migration.GetID()))
	pathgz := path + ".tar.gz"

	err = downloadURL(url, pathgz)
	if err != nil {
		return err
	}

	elapsed = time.Since(t1)
	logger.With(log.Fields{"elapsed": elapsed}).Infof("file downloaded to %v", pathgz)
	t2 := time.Now()

	err = archiver.Unarchive(pathgz, path)
	if err != nil {
		return err
	}

	elapsed = time.Since(t2)
	logger.With(log.Fields{"elapsed": elapsed}).Infof("file uncompressed in %v", path)

	elapsed = time.Since(t0)
	logger.With(log.Fields{"total-elapsed": elapsed}).Infof("done")

	return nil
}

func (d GitHubMigrationDownloader) DownloadOrg(name string, version string) error {
	return fmt.Errorf("not implemented")
}

func (d GitHubMigrationDownloader) SetCurrent(version string) error {
	return fmt.Errorf("not implemented")
}

func downloadURL(url, dst string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status %v", resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
