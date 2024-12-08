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

### Development
Make changes and run tests
```
make tests
```
Clean up containers and docker network
```
make clean
```

### Caveats

This tool only includes a naive file parser, if one can call it that.
The format `/* rollback\n ...\n*/` must be followed otherwise rollback SQL statements will not be parsed correctly and may cause undefined behavior.

This tool does not provide a way to track changes to applied migrations. Only, that it is expected to return an error when the contents of a migration file diverges from a previously run migration file in the same filesystem "order". It is up to the user to track these changes via a version control system such as git.

As the filesystem migration files and the database version table are both considered sources of truth, users are encouraged to prefix migration files with their intended version `001_add_users_table.sql` and to avoid skipping version numbers to mirror version state in the database and avoid discrepancies. This tool does not allow missing migrations to be applied on subsequent runs, eg. a new `005_add_some_column.sql` will not be applied when the current version recorded on the DB is a later version. In that scenario, users may opt to rollback to an older version (if the business case allows for it) and then reapply later versions.
