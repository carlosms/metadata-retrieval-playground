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

const (
	repositoriesCols  = "database_id, created_at, description, owner, name"
	issuesCols        = "database_id, title, body, number, repository_owner, repository_name"
	issueCommentsCols = "database_id, author, body, repository_owner, repository_name, issue_number"
)

func (s *dbStorer) setActiveVersion(v string) error {
	// TODO: for some reason the normal parameter interpolation $1 fails with
	// pq: got 1 parameters but the statement requires 0

	_, err := s.db.Exec(fmt.Sprintf(`CREATE OR REPLACE VIEW repositories AS
	SELECT %s
	FROM repositories_versioned WHERE '%s' = ANY(versions)`, repositoriesCols, v))
	if err != nil {
		return err
	}

	_, err = s.db.Exec(fmt.Sprintf(`CREATE OR REPLACE VIEW issues AS
	SELECT %s
	FROM issues_versioned WHERE '%s' = ANY(versions)`, issuesCols, v))
	if err != nil {
		return err
	}

	_, err = s.db.Exec(fmt.Sprintf(`CREATE OR REPLACE VIEW issue_comments AS
	SELECT %s
	FROM issue_comments_versioned WHERE '%s' = ANY(versions)`, issueCommentsCols, v))
	if err != nil {
		return err
	}

	return nil
}

func (s *dbStorer) cleanup(currentVersion string) error {
	tables := []string{"repositories_versioned", "issues_versioned", "issue_comments_versioned"}

	for _, table := range tables {
		// Delete all entries that do not belong to currentVersion
		_, err := s.db.Exec(fmt.Sprintf(`DELETE FROM %s WHERE '%s' <> ALL(versions)`, table, currentVersion))
		if err != nil {
			return err
		}

		// All remaining entries belong to currentVersion, replace the list of versions
		// with an array of 1 entry
		_, err = s.db.Exec(fmt.Sprintf(`UPDATE %s SET versions = array['%s']`, table, currentVersion))
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *dbStorer) saveRepository(repository *RepositoryFields) error {
	statement := fmt.Sprintf(
		`INSERT INTO repositories_versioned
		(versions, %s)
		VALUES (array[$1], $2, $3, $4, $5, $6)
		ON CONFLICT (%s)
		DO UPDATE
		SET versions = array_append(repositories_versioned.versions, $1)`,
		repositoriesCols, repositoriesCols)

	_, err := s.tx.Exec(statement,
		s.v,
		repository.DatabaseId, repository.CreatedAt, repository.Description,
		repository.Owner.Login, repository.Name)

	return err
}

func (s *dbStorer) saveIssue(repositoryOwner, repositoryName string, issue *Issue) error {
	statement := fmt.Sprintf(
		`INSERT INTO issues_versioned
		(versions, %s)
		VALUES (array[$1], $2, $3, $4, $5, $6, $7)
		ON CONFLICT (%s)
		DO UPDATE
		SET versions = array_append(issues_versioned.versions, $1)`,
		issuesCols, issuesCols)

	_, err := s.tx.Exec(statement,
		s.v,
		issue.DatabaseId, issue.Title, issue.Body, issue.Number,
		repositoryOwner, repositoryName)

	return err
}

func (s *dbStorer) saveIssueComment(repositoryOwner, repositoryName string, issueNumber int, comment *IssueComment) error {
	statement := fmt.Sprintf(`INSERT INTO issue_comments_versioned
		(versions, %s)
		VALUES (array[$1], $2, $3, $4, $5, $6, $7)
		ON CONFLICT (%s)
		DO UPDATE
		SET versions = array_append(issue_comments_versioned.versions, $1)`,
		issueCommentsCols, issueCommentsCols)

	_, err := s.tx.Exec(statement,
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
