package v3

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/carlosms/metadata-retrieval-playground"
	"github.com/carlosms/metadata-retrieval-playground/internal/client"
	"github.com/carlosms/metadata-retrieval-playground/internal/store"
	"github.com/google/go-github/github"
	"gopkg.in/src-d/go-log.v1"
)

type GitHubDownloader struct {
	store.Storer

	client *github.Client
}

var _ metadata.MetadataDownloader = GitHubDownloader{}

func NewStdoutDownloader(httpClient *http.Client) (*GitHubDownloader, error) {
	c, err := client.NewClient(httpClient)
	if err != nil {
		return nil, err
	}

	return &GitHubDownloader{
		Storer: store.StdoutStorer{},
		client: github.NewClient(c),
	}, nil
}

func (d GitHubDownloader) DownloadRepository(owner string, name string, version string) error {
	logger := log.New(log.Fields{"owner": owner, "repo": name})

	rate0, err := d.rateRemaining()
	if err != nil {
		return err
	}

	t0 := time.Now()

	repository, _, err := d.client.Repositories.Get(context.TODO(), owner, name)
	if err != nil {
		return err
	}

	err = d.SaveRepository(repository)
	if err != nil {
		return err
	}

	elapsed := time.Since(t0)
	logger.With(log.Fields{"elapsed": elapsed}).Infof("repository metadata fetched")

	t1 := time.Now()

	// issues, PRs and comments
	err = d.downloadIssues(logger, owner, name, version)
	if err != nil {
		return err
	}

	elapsed = time.Since(t1)
	logger.With(log.Fields{"elapsed": elapsed}).Infof("issues & issue comments fetched")

	elapsed = time.Since(t0)
	rate1, err := d.rateRemaining()
	if err != nil {
		return err
	}
	rateUsed := rate0 - rate1

	logger.With(log.Fields{"rate-limit-used": rateUsed, "total-elapsed": elapsed}).Infof("All metadata fetched")

	return nil
}

func (d GitHubDownloader) rateRemaining() (int, error) {
	limit, _, err := d.client.RateLimits(context.TODO())
	if err != nil {
		return 0, err
	}

	return limit.GetCore().Remaining, nil
}

const listOptionsPerPage = 100

func (d GitHubDownloader) downloadIssues(logger log.Logger, owner string, repo string, version string) error {
	opts := &github.IssueListByRepoOptions{}
	opts.ListOptions.PerPage = listOptionsPerPage
	opts.State = "all"

	for {
		issues, r, err := d.client.Issues.ListByRepo(context.TODO(), owner, repo, opts)
		if err != nil {
			return err
		}

		// TODO: there is an endpoint to get all repo issue comments,
		// instead of calling for each issue/PR

		for _, i := range issues {
			if i.IsPullRequest() {
				err := d.downloadPullRequest(owner, repo, i.GetNumber(), version)
				if err != nil {
					return err
				}
				// PRs have: normal issue comments, individual review comments, and Reviews (groups of Review comments)
				err = d.downloadIssueComments(owner, repo, i.GetNumber(), version)
				if err != nil {
					return err
				}

				err = d.downloadPullRequestReviews(owner, repo, i.GetNumber(), version)
				if err != nil {
					return err
				}

				err = d.downloadPullRequestComments(owner, repo, i.GetNumber(), version)
				if err != nil {
					return err
				}
			} else {
				err := d.downloadIssue(owner, repo, i.GetNumber(), version)
				if err != nil {
					return err
				}
				err = d.downloadIssueComments(owner, repo, i.GetNumber(), version)
				if err != nil {
					return err
				}
			}
		}

		if r.NextPage == 0 {
			break
		}

		opts.Page = r.NextPage
	}

	return nil
}

func (d GitHubDownloader) downloadIssue(owner string, repo string, number int, version string) error {
	issue, _, err := d.client.Issues.Get(context.TODO(), owner, repo, number)
	if err != nil {
		return err
	}

	return d.SaveIssue(issue)
}

func (d GitHubDownloader) downloadIssueComments(owner string, repo string, number int, version string) error {
	opts := &github.IssueListCommentsOptions{}
	opts.ListOptions.PerPage = listOptionsPerPage

	for {
		comments, r, err := d.client.Issues.ListComments(context.TODO(), owner, repo, number, opts)
		if err != nil {
			return err
		}

		for _, comment := range comments {
			// No need to do
			// d.client.Issues.GetComment(context.TODO(), owner, repo, comment.GetID())
			// the contents are the same, see
			// https://developer.github.com/v3/issues/comments/#get-a-single-comment
			// https://developer.github.com/v3/issues/comments/#list-comments-on-an-issue

			err = d.SaveIssueComment(comment)
			if err != nil {
				return err
			}
		}

		if r.NextPage == 0 {
			break
		}

		opts.Page = r.NextPage
	}

	return nil
}

func (d GitHubDownloader) downloadPullRequest(owner string, repo string, number int, version string) error {
	issue, _, err := d.client.PullRequests.Get(context.TODO(), owner, repo, number)
	if err != nil {
		return err
	}

	return d.SavePullRequest(issue)
}

func (d GitHubDownloader) downloadPullRequestReviews(owner string, repo string, number int, version string) error {
	opts := &github.ListOptions{}
	opts.PerPage = listOptionsPerPage

	for {
		reviews, r, err := d.client.PullRequests.ListReviews(context.TODO(), owner, repo, number, opts)
		if err != nil {
			return err
		}

		for _, review := range reviews {
			// No need to do
			// d.client.Issues.GetComment(context.TODO(), owner, repo, comment.GetID())
			// the contents are the same, see
			// https://developer.github.com/v3/issues/comments/#get-a-single-comment
			// https://developer.github.com/v3/issues/comments/#list-comments-on-an-issue

			err = d.SavePullRequestReview(review)
			if err != nil {
				return err
			}
		}

		if r.NextPage == 0 {
			break
		}

		opts.Page = r.NextPage
	}

	return nil
}

func (d GitHubDownloader) downloadPullRequestComments(owner string, repo string, number int, version string) error {
	opts := &github.PullRequestListCommentsOptions{}
	opts.ListOptions.PerPage = listOptionsPerPage

	for {
		comments, r, err := d.client.PullRequests.ListComments(context.TODO(), owner, repo, number, opts)
		if err != nil {
			return err
		}

		for _, comment := range comments {
			// No need to do
			// d.client.Issues.GetComment(context.TODO(), owner, repo, comment.GetID())
			// the contents are the same, see
			// https://developer.github.com/v3/issues/comments/#get-a-single-comment
			// https://developer.github.com/v3/issues/comments/#list-comments-on-an-issue

			err = d.SavePullRequestComment(comment)
			if err != nil {
				return err
			}
		}

		if r.NextPage == 0 {
			break
		}

		opts.Page = r.NextPage
	}

	return nil
}

func (d GitHubDownloader) DownloadOrg(name string, version string) error {
	return fmt.Errorf("not implemented")
}

func (d GitHubDownloader) SetCurrent(version string) error {
	return fmt.Errorf("not implemented")
}
