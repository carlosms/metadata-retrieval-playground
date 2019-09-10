BEGIN;

CREATE TABLE IF NOT EXISTS repositories (
  database_id  integer PRIMARY KEY,
  created_at   timestamptz,
  description  text,
  owner        text,
  name text,
  UNIQUE (owner, name)
);

CREATE TABLE IF NOT EXISTS issues (
  database_id  integer PRIMARY KEY,
  title text,
  body text,
  number int,
  repository_owner text,
  repository_name text,
  UNIQUE (repository_owner, repository_name, number)
);

CREATE TABLE IF NOT EXISTS issue_comments (
  database_id  integer PRIMARY KEY,
  author text,
  body text,
  repository_owner text,
  repository_name text,
  issue_number int
);

COMMIT;
