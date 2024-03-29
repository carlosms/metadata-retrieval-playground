package v4

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/carlosms/metadata-retrieval-playground"
	"github.com/carlosms/metadata-retrieval-playground/internal/client"
	"github.com/shurcooL/githubv4"
	"gopkg.in/src-d/go-log.v1"
)

// max pageList is 100. But with nested fields, github may return error:
// By the time this query traverses to the comments connection, it is requesting up to 1,000,000 possible nodes which exceeds the maximum limit of 500,000.
// e.g. 100*100*100
// {
//   repository(owner: "carlosms-test-org", name: "test-repo") {
// 		pullRequests(first:100){
//       nodes{
//         reviews(first:100){
//           nodes{
//             comments(first:100){
//               totalCount
//             }
//           }
//         }
//       }
//     }
//   }
// }
const pageList = 40

type storer interface {
	saveRepository(repository *RepositoryFields) error
	saveIssue(repositoryOwner, repositoryName string, issue *Issue) error
	saveIssueComment(repositoryOwner, repositoryName string, issueNumber int, comment *IssueComment) error
	savePullRequest(pr *PullRequest) error
	savePullRequestReview(review *PullRequestReview) error
	saveReviewComment(comment *PullRequestReviewComment) error

	begin() error
	commit() error
	rollback() error
	version(v string)
	setActiveVersion(v string) error
	cleanup(currentVersion string) error
}

type GitHubDownloader struct {
	storer

	client *githubv4.Client
}

var _ metadata.MetadataDownloader = GitHubDownloader{}

func NewStdoutDownloader(httpClient *http.Client) (*GitHubDownloader, error) {
	c, err := client.NewClient(httpClient)
	if err != nil {
		return nil, err
	}

	return &GitHubDownloader{
		storer: &stdoutStorer{},
		client: githubv4.NewClient(c),
	}, nil
}

func NewDBDownloader(httpClient *http.Client, db *sql.DB) (*GitHubDownloader, error) {
	c, err := client.NewClient(httpClient)
	if err != nil {
		return nil, err
	}

	return &GitHubDownloader{
		storer: &dbStorer{db: db},
		client: githubv4.NewClient(c),
	}, nil
}

func (d GitHubDownloader) DownloadRepository(owner string, name string, version string) error {
	logger := log.New(log.Fields{"owner": owner, "repo": name})

	d.storer.version(version)

	var err error
	err = d.storer.begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			d.storer.rollback()
			return
		}

		d.storer.commit()
	}()

	rate0, err := d.rateRemaining()
	if err != nil {
		return err
	}
	t0 := time.Now()

	var q struct {
		Repository `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":                           githubv4.String(owner),
		"name":                            githubv4.String(name),
		"pageList":                        githubv4.Int(pageList),
		"issuesCursor":                    (*githubv4.String)(nil),
		"issueCommentsCursor":             (*githubv4.String)(nil),
		"pullRequestsCursor":              (*githubv4.String)(nil),
		"pullRequestReviewsCursor":        (*githubv4.String)(nil),
		"pullRequestReviewCommentsCursor": (*githubv4.String)(nil),
	}

	err = d.client.Query(context.TODO(), &q, variables)
	if err != nil {
		return err
	}

	err = d.storer.saveRepository(&q.Repository.RepositoryFields)
	if err != nil {
		return err
	}

	elapsed := time.Since(t0)
	logger.With(log.Fields{"elapsed": elapsed}).Infof("repository metadata fetched")

	t1 := time.Now()

	// issues and comments
	err = d.downloadIssues(logger, owner, name, &q.Repository)
	if err != nil {
		return err
	}

	elapsed = time.Since(t1)
	logger.With(log.Fields{"elapsed": elapsed}).Infof("issues & issue comments fetched")

	t2 := time.Now()

	// PRs and comments
	err = d.downloadPullRequests(logger, owner, name, &q.Repository)
	if err != nil {
		return err
	}

	elapsed = time.Since(t2)
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
	var q struct {
		RateLimit struct {
			Remaining int
		}
	}

	err := d.client.Query(context.TODO(), &q, nil)
	if err != nil {
		return 0, err
	}

	return q.RateLimit.Remaining, nil
}

func (d GitHubDownloader) downloadIssues(logger log.Logger, owner string, name string, repository *Repository) error {
	process := func(issue *Issue) error {
		err := d.storer.saveIssue(owner, name, issue)
		if err != nil {
			return err
		}
		return d.downloadIssueComments(logger.With(log.Fields{"issue": issue.Number}), owner, name, issue)
	}

	// Save issues included in the first page
	for _, issue := range repository.Issues.Nodes {
		err := process(&issue)
		if err != nil {
			return err
		}
	}

	variables := map[string]interface{}{
		"owner":               githubv4.String(owner),
		"name":                githubv4.String(name),
		"pageList":            githubv4.Int(pageList),
		"issueCommentsCursor": (*githubv4.String)(nil),
	}

	// if there are more issues, loop over all the pages
	hasNextPage := repository.Issues.PageInfo.HasNextPage
	endCursor := repository.Issues.PageInfo.EndCursor

	for hasNextPage {
		logger.Debugf("issues loop")

		// get only issues
		var q struct {
			Repository struct {
				Issues IssueConnection `graphql:"issues(first: $pageList, after: $issuesCursor)"`
			} `graphql:"repository(owner: $owner, name: $name)"`
		}

		variables["issuesCursor"] = githubv4.String(endCursor)

		err := d.client.Query(context.TODO(), &q, variables)
		if err != nil {
			return err
		}

		for _, issue := range q.Repository.Issues.Nodes {
			err := process(&issue)
			if err != nil {
				return err
			}
		}

		hasNextPage = q.Repository.Issues.PageInfo.HasNextPage
		endCursor = q.Repository.Issues.PageInfo.EndCursor
	}

	return nil
}

func (d GitHubDownloader) downloadIssueComments(logger log.Logger, owner string, name string, issue *Issue) error {
	// save first page of comments
	for _, comment := range issue.Comments.Nodes {
		err := d.storer.saveIssueComment(owner, name, issue.Number, &comment)
		if err != nil {
			return err
		}
	}

	variables := map[string]interface{}{
		"owner":       githubv4.String(owner),
		"name":        githubv4.String(name),
		"pageList":    githubv4.Int(pageList),
		"issueNumber": githubv4.Int(issue.Number),
	}

	// if there are more issue comments, loop over all the pages
	hasNextPage := issue.Comments.PageInfo.HasNextPage
	endCursor := issue.Comments.PageInfo.EndCursor

	for hasNextPage {
		logger.Debugf("issue comments loop")

		// get only issue comments
		var q struct {
			Repository struct {
				Issue struct {
					Comments IssueCommentsConnection `graphql:"comments(first: $pageList, after: $issueCommentsCursor)"`
				} `graphql:"issue(number: $issueNumber)"`
			} `graphql:"repository(owner: $owner, name: $name)"`
		}

		variables["issueCommentsCursor"] = githubv4.String(endCursor)

		err := d.client.Query(context.TODO(), &q, variables)
		if err != nil {
			return err
		}

		for _, comment := range q.Repository.Issue.Comments.Nodes {
			err := d.storer.saveIssueComment(owner, name, issue.Number, &comment)
			if err != nil {
				return err
			}
		}

		hasNextPage = q.Repository.Issue.Comments.PageInfo.HasNextPage
		endCursor = q.Repository.Issue.Comments.PageInfo.EndCursor
	}

	return nil
}

func (d GitHubDownloader) downloadPullRequests(logger log.Logger, owner string, name string, repository *Repository) error {
	process := func(pr *PullRequest) error {
		err := d.storer.savePullRequest(pr)
		if err != nil {
			return err
		}
		err = d.downloadPullRequestComments(logger.With(log.Fields{"pr": pr.Number}), owner, name, pr)
		if err != nil {
			return err
		}
		err = d.downloadPullRequestReviews(logger.With(log.Fields{"pr": pr.Number}), owner, name, pr)
		if err != nil {
			return err
		}

		return nil
	}

	// Save PRs included in the first page
	for _, pr := range repository.PullRequests.Nodes {
		err := process(&pr)
		if err != nil {
			return err
		}
	}

	variables := map[string]interface{}{
		"owner":                           githubv4.String(owner),
		"name":                            githubv4.String(name),
		"pageList":                        githubv4.Int(pageList),
		"issueCommentsCursor":             (*githubv4.String)(nil),
		"pullRequestReviewsCursor":        (*githubv4.String)(nil),
		"pullRequestReviewCommentsCursor": (*githubv4.String)(nil),
	}

	// if there are more PRs, loop over all the pages
	hasNextPage := repository.PullRequests.PageInfo.HasNextPage
	endCursor := repository.PullRequests.PageInfo.EndCursor

	for hasNextPage {
		logger.Debugf("PRs loop")

		// get only PRs
		var q struct {
			Repository struct {
				PullRequests PullRequestConnection `graphql:"pullRequests(first: $pageList, after: $pullRequestsCursor)"`
			} `graphql:"repository(owner: $owner, name: $name)"`
		}

		variables["pullRequestsCursor"] = githubv4.String(endCursor)

		err := d.client.Query(context.TODO(), &q, variables)
		if err != nil {
			return err
		}

		for _, pr := range q.Repository.PullRequests.Nodes {
			err := process(&pr)
			if err != nil {
				return err
			}
		}

		hasNextPage = q.Repository.PullRequests.PageInfo.HasNextPage
		endCursor = q.Repository.PullRequests.PageInfo.EndCursor
	}

	return nil
}

func (d GitHubDownloader) downloadPullRequestComments(logger log.Logger, owner string, name string, pr *PullRequest) error {
	// save first page of comments
	for _, comment := range pr.Comments.Nodes {
		err := d.storer.saveIssueComment(owner, name, pr.Number, &comment)
		if err != nil {
			return err
		}
	}

	variables := map[string]interface{}{
		"owner":       githubv4.String(owner),
		"name":        githubv4.String(name),
		"pageList":    githubv4.Int(pageList),
		"issueNumber": githubv4.Int(pr.Number),
	}

	// if there are more issue comments, loop over all the pages
	hasNextPage := pr.Comments.PageInfo.HasNextPage
	endCursor := pr.Comments.PageInfo.EndCursor

	for hasNextPage {
		logger.Debugf("PR comments loop")

		// get only PR comments
		var q struct {
			Repository struct {
				PullRequest struct {
					Comments IssueCommentsConnection `graphql:"comments(first: $pageList, after: $issueCommentsCursor)"`
				} `graphql:"issue(number: $issueNumber)"`
			} `graphql:"repository(owner: $owner, name: $name)"`
		}

		variables["issueCommentsCursor"] = githubv4.String(endCursor)

		err := d.client.Query(context.TODO(), &q, variables)
		if err != nil {
			return err
		}

		for _, comment := range q.Repository.PullRequest.Comments.Nodes {
			err := d.storer.saveIssueComment(owner, name, pr.Number, &comment)
			if err != nil {
				return err
			}
		}

		hasNextPage = q.Repository.PullRequest.Comments.PageInfo.HasNextPage
		endCursor = q.Repository.PullRequest.Comments.PageInfo.EndCursor
	}

	return nil
}

func (d GitHubDownloader) downloadPullRequestReviews(logger log.Logger, owner string, name string, pr *PullRequest) error {
	process := func(review *PullRequestReview) error {
		err := d.storer.savePullRequestReview(review)
		if err != nil {
			return err
		}
		return d.downloadReviewComments(logger.With(log.Fields{"pr": pr.Number}), owner, name, pr.Number, review)
	}

	// save first page of reviews
	for _, review := range pr.Reviews.Nodes {
		err := process(&review)
		if err != nil {
			return err
		}
	}

	variables := map[string]interface{}{
		"owner":                           githubv4.String(owner),
		"name":                            githubv4.String(name),
		"pageList":                        githubv4.Int(pageList),
		"issueNumber":                     githubv4.Int(pr.Number),
		"pullRequestReviewsCursor":        (*githubv4.String)(nil),
		"pullRequestReviewCommentsCursor": (*githubv4.String)(nil),
	}

	// if there are more reviews, loop over all the pages
	hasNextPage := pr.Reviews.PageInfo.HasNextPage
	endCursor := pr.Reviews.PageInfo.EndCursor

	for hasNextPage {
		logger.Debugf("PR reviews loop")

		// get only PR reviews
		var q struct {
			Repository struct {
				PullRequest struct {
					Reviews PullRequestReviewConnection `graphql:"reviews(first: $pageList, after: $pullRequestReviewsCursor)"`
				} `graphql:"pullRequest(number: $issueNumber)"`
			} `graphql:"repository(owner: $owner, name: $name)"`
		}

		variables["pullRequestReviewsCursor"] = githubv4.String(endCursor)

		err := d.client.Query(context.TODO(), &q, variables)
		if err != nil {
			return err
		}

		for _, review := range q.Repository.PullRequest.Reviews.Nodes {
			process(&review)
		}

		hasNextPage = q.Repository.PullRequest.Reviews.PageInfo.HasNextPage
		endCursor = q.Repository.PullRequest.Reviews.PageInfo.EndCursor
	}

	return nil
}

func (d GitHubDownloader) downloadReviewComments(logger log.Logger, repositoryOwner, repositoryName string, issueNumber int, review *PullRequestReview) error {
	// save first page of comments
	for _, comment := range review.Comments.Nodes {
		err := d.storer.saveReviewComment(&comment)
		if err != nil {
			return err
		}
	}

	/*
		TODO: how to perform pagination at this level

		There isn't a way to ask for repository/pullRequest(number:3)/review(id:X),
		you can only query all reviews

		{
		  repository(owner: "carlosms-test-org", name: "test-repo") {
		    pullRequest(number:3) {
		      reviews {.....}
		    }
		  }
		}
	*/
	if review.Comments.PageInfo.HasNextPage {
		log.Errorf(nil, "PR review contains more than on page of comments, but pagination is not implemented")
	}

	return nil
}

func (d GitHubDownloader) DownloadOrg(name string, version string) error {
	return fmt.Errorf("not implemented")
}

func (d GitHubDownloader) SetCurrent(version string) error {
	return d.storer.setActiveVersion(version)
}

// Cleanup deletes from the DB all records that do not belong to the currentVersion
func (d GitHubDownloader) Cleanup(currentVersion string) error {
	return d.storer.cleanup(currentVersion)
}
