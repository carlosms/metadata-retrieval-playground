package v4

import (
	"database/sql"

	"gopkg.in/src-d/go-log.v1"
)

type dbStorer struct {
	db *sql.DB
	tx *sql.Tx
}

func (s *dbStorer) begin() error {
	var err error
	s.tx, err = s.db.Begin()
	return err
}

func (s *dbStorer) commit() error {
	return s.tx.Commit()
}

func (s *dbStorer) rollback() error {
	return s.tx.Rollback()
}

func (s *dbStorer) saveRepository(repository *RepositoryFields) error {
	_, err := s.tx.Exec(`INSERT INTO repositories
		(database_id, created_at, description, owner, name)
		VALUES ($1, $2, $3, $4, $5)`,
		repository.DatabaseId, repository.CreatedAt, repository.Description,
		repository.Owner.Login, repository.Name)

	return err
}

func (s *dbStorer) saveIssue(repositoryOwner, repositoryName string, issue *Issue) error {
	_, err := s.tx.Exec(`INSERT INTO issues
		(database_id, title, body, number, repository_owner, repository_name)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		issue.DatabaseId, issue.Title, issue.Body, issue.Number,
		repositoryOwner, repositoryName)

	return err
}

func (s *dbStorer) saveIssueComment(repositoryOwner, repositoryName string, issueNumber int, comment *IssueComment) error {
	_, err := s.tx.Exec(`INSERT INTO issue_comments
		(database_id, author, body, repository_owner, repository_name, issue_number)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		comment.DatabaseId, comment.Author.Login, comment.Body,
		repositoryName, repositoryOwner, issueNumber)

	return err
}

func (s *dbStorer) savePullRequest(pr *PullRequest) error {
	log.Errorf(nil, "savePullRequest not implemented")
	return nil
}

func (s *dbStorer) savePullRequestReview(review *PullRequestReview) error {
	log.Errorf(nil, "savePullRequestReview not implemented")
	return nil
}

func (s *dbStorer) saveReviewComment(comment *PullRequestReviewComment) error {
	log.Errorf(nil, "saveReviewComment not implemented")
	return nil
}
