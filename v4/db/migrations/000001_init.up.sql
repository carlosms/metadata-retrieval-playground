BEGIN;

CREATE TABLE IF NOT EXISTS repositories_versioned (
  pk           SERIAL PRIMARY KEY,
  versions     text ARRAY,
  
  database_id  integer,
  created_at   timestamptz,
  description  text,
  owner        text,
  name         text,
  
  UNIQUE(database_id, created_at, description, owner, name)
);

CREATE INDEX IF NOT EXISTS versions ON repositories_versioned (versions);

CREATE TABLE IF NOT EXISTS issues_versioned (
  pk                SERIAL PRIMARY KEY,
  versions          text ARRAY,

  database_id       integer,
  title             text,
  body              text,
  number            int,
  repository_owner  text,
  repository_name   text,

  UNIQUE(database_id, title, body, number, repository_owner, repository_name)
);

CREATE INDEX IF NOT EXISTS versions ON issues_versioned (versions);

CREATE TABLE IF NOT EXISTS issue_comments_versioned (
  pk                SERIAL PRIMARY KEY,
  versions          text ARRAY,

  database_id       integer,
  author            text,
  body              text,
  repository_owner  text,
  repository_name   text,
  issue_number      int,

  UNIQUE(database_id, author, body, repository_owner, repository_name, issue_number)
);

CREATE INDEX IF NOT EXISTS versions ON issue_comments_versioned (versions);

COMMIT;
