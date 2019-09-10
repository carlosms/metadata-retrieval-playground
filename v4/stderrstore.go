package v4

import (
	"fmt"
)

type stdoutStorer struct{}

func (s *stdoutStorer) saveRepository(repository *RepositoryFields) error {
	fmt.Printf("repository data fetched for %s/%s\n", repository.Owner.Login, repository.Name)
	return nil
}

func (s *stdoutStorer) saveIssue(repositoryOwner, repositoryName string, issue *Issue) error {
	fmt.Printf("issue data fetched for #%v %s\n", issue.Number, issue.Title)
	return nil
}

func (s *stdoutStorer) saveIssueComment(repositoryOwner, repositoryName string, issueNumber int, comment *IssueComment) error {
	fmt.Printf("  issue comment data fetched by %s at %v: %q\n", comment.Author.Login, comment.CreatedAt, trim(comment.Body))
	return nil
}

func (s *stdoutStorer) savePullRequest(pr *PullRequest) error {
	fmt.Printf("PR data fetched for #%v %s\n", pr.Number, pr.Title)
	return nil
}

func (s *stdoutStorer) savePullRequestReview(review *PullRequestReview) error {
	fmt.Printf("  PR Review data fetched by %s at %v: %q\n", review.Author.Login, review.CreatedAt, trim(review.Body))
	return nil
}

func (s *stdoutStorer) saveReviewComment(comment *PullRequestReviewComment) error {
	fmt.Printf("    PR review comment data fetched by %s at %v: %q\n", comment.Author.Login, comment.CreatedAt, trim(comment.Body))
	return nil
}

func (s *stdoutStorer) begin() error {
	return nil
}

func (s *stdoutStorer) commit() error {
	return nil
}

func (s *stdoutStorer) rollback() error {
	return nil
}

func trim(s string) string {
	if len(s) > 40 {
		return s[0:39] + "..."
	}

	return s
}
