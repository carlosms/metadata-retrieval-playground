package v4

import (
	"database/sql"
	"fmt"

	"gopkg.in/src-d/go-log.v1"
)

type dbStorer struct {
	db *sql.DB
	tx *sql.Tx
	v  string
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

func (s *dbStorer) version(v string) {
	s.v = v
}

func (s *dbStorer) setActiveVersion(v string) error {
	// TODO: for some reason the normal parameter interpolation $1 fails with
	// pq: got 1 parameters but the statement requires 0

	_, err := s.db.Exec(fmt.Sprintf(`CREATE OR REPLACE VIEW repositories AS
	SELECT database_id, created_at, description, owner, name
	FROM repositories_versioned WHERE '%s' = ANY(versions)`, v))
	if err != nil {
		return err
	}

	_, err = s.db.Exec(fmt.Sprintf(`CREATE OR REPLACE VIEW issues AS
	SELECT database_id, title, body, number, repository_owner, repository_name
	FROM issues_versioned WHERE '%s' = ANY(versions)`, v))
	if err != nil {
		return err
	}

	_, err = s.db.Exec(fmt.Sprintf(`CREATE OR REPLACE VIEW issue_comments AS
	SELECT database_id, author, body, repository_owner, repository_name, issue_number
	FROM issue_comments_versioned WHERE '%s' = ANY(versions)`, v))
	if err != nil {
		return err
	}

	return nil
}

func (s *dbStorer) saveRepository(repository *RepositoryFields) error {
	_, err := s.tx.Exec(
		`INSERT INTO repositories_versioned
		(versions, database_id, created_at, description, owner, name)
		VALUES (array[$1], $2, $3, $4, $5, $6)
		ON CONFLICT (database_id, created_at, description, owner, name)
		DO UPDATE
		SET versions = array_append(repositories_versioned.versions, $1)`,
		s.v,
		repository.DatabaseId, repository.CreatedAt, repository.Description,
		repository.Owner.Login, repository.Name)

	return err
}

func (s *dbStorer) saveIssue(repositoryOwner, repositoryName string, issue *Issue) error {
	_, err := s.tx.Exec(
		`INSERT INTO issues_versioned
		(versions, database_id, title, body, number, repository_owner, repository_name)
		VALUES (array[$1], $2, $3, $4, $5, $6, $7)
		ON CONFLICT (database_id, title, body, number, repository_owner, repository_name)
		DO UPDATE
		SET versions = array_append(issues_versioned.versions, $1)`,
		s.v,
		issue.DatabaseId, issue.Title, issue.Body, issue.Number,
		repositoryOwner, repositoryName)

	return err
}

func (s *dbStorer) saveIssueComment(repositoryOwner, repositoryName string, issueNumber int, comment *IssueComment) error {
	_, err := s.tx.Exec(
		`INSERT INTO issue_comments_versioned
		(versions, database_id, author, body, repository_owner, repository_name, issue_number)
		VALUES (array[$1], $2, $3, $4, $5, $6, $7)
		ON CONFLICT (database_id, author, body, repository_owner, repository_name, issue_number)
		DO UPDATE
		SET versions = array_append(issue_comments_versioned.versions, $1)`,
		s.v,
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
