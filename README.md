mitch
---
A simple (maybe) migration tool for Clickhouse.

Yet another coding challenge submission. ;)

## Migrations

A sample SQL migration file content looks like:

```sql
CREATE TABLE post (
    id int NOT NULL,
    title text,
    body text,
    PRIMARY KEY(id)
);

/* rollback
DROP TABLE post;
*/
```

This tool only includes a naive file parser, if one can call it that.
The format `/* rollback\n ...\n*/` must be followed otherwise rollback SQL statements will not be parsed correctly and may cause undefined behavior.

This tool does not provide a way to track changes to applied migrations. Only, that it is expected to return an error when the contents of a migration file diverges from a previously run migration file in the same filesystem "order". It is up to the user to track these changes via a version control system such as git.

As the filesystem and the database version table are considered both sources of truth. It is highly encouraged to prefix migration files with their intended version and to avoid skipping version numbers, eg. `001_add_users_table.sql` to mirror version state in the database and avoid discrepancies.