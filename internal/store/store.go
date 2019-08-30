package store

import (
	"fmt"

	"github.com/google/go-github/github"
)

type Storer interface {
	SaveRepository(repository *github.Repository) error
	SaveIssue(issue *github.Issue) error
	SaveIssueComment(comment *github.IssueComment) error
	SavePullRequest(issue *github.PullRequest) error
	SavePullRequestComment(comment *github.PullRequestComment) error
	SavePullRequestReview(comment *github.PullRequestReview) error
}

type StdoutStorer struct{}

func (s StdoutStorer) SaveRepository(repository *github.Repository) error {
	fmt.Printf("repository data fetched for %s/%s\n", repository.GetOwner().GetName(), repository.GetName())
	return nil
}

func (s StdoutStorer) SaveIssue(issue *github.Issue) error {
	fmt.Printf("issue data fetched for #%v %s\n", issue.GetNumber(), issue.GetTitle())
	return nil
}

func (s StdoutStorer) SaveIssueComment(comment *github.IssueComment) error {
	fmt.Printf("  issue comment data fetched %v by %s: %q\n", comment.GetIssueURL(), comment.GetUser().GetLogin(), trim(comment.GetBody()))
	return nil
}

func (s StdoutStorer) SavePullRequest(pr *github.PullRequest) error {
	fmt.Printf("PR data fetched for #%v %s\n", pr.GetNumber(), pr.GetTitle())
	return nil
}

func (s StdoutStorer) SavePullRequestComment(comment *github.PullRequestComment) error {
	fmt.Printf("  PR comment data fetched %v by %s: %q\n", comment.GetPullRequestURL(), comment.GetUser().GetLogin(), trim(comment.GetBody()))
	return nil
}

func (s StdoutStorer) SavePullRequestReview(review *github.PullRequestReview) error {
	fmt.Printf("  PR review data fetched %v by %s %v: %q\n", review.GetPullRequestURL(), review.GetUser().GetLogin(), review.GetState(), trim(review.GetBody()))
	return nil
}

func trim(s string) string {
	if len(s) > 40 {
		return s[0:39] + "..."
	}

	return s
}
