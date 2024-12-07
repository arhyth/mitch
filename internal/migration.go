package internal

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	"golang.org/x/sync/errgroup"
)

type Version struct {
	Id       int
	Up, Down *Migration
}

type Migration struct {
	SQL string
}

type Migrations []*Version

func (ms Migrations) FillVersionAtIndex(idx int, file fs.File) error {
	ver, err := ParseMigration(file)
	if err != nil {
		return err
	}
	ms[idx] = ver
	return nil
}

type Runner struct {
	Dir fs.FS
	DB  *sql.DB
}

// TODO
func (rr *Runner) Migrate() error {
	return nil
}

// TODO
func (rr *Runner) RollbackTo(fname string) error {
	return nil
}

func (rr *Runner) mustVersionTable() error {
	dbname := rr.getDBName()
	var query = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.mitch_db_version (
		    version_id Int64,
		    source String,
			content_hash FixedString(32),
		    created_at DateTime default now()
		)
		ENGINE = MergeTree()
		ORDER BY (version_id, content_hash);
	`, dbname)
	_, err := rr.DB.ExecContext(context.Background(), query)
	if err != nil {
		return err
	}

	return nil
}

func (rr *Runner) getDBName() string {
	query := "SELECT currentDatabase();"
	var dbName string
	err := rr.DB.QueryRow(query).Scan(&dbName)
	if err != nil {
		// panic here since the database connection setup in `Connect`
		// is supposed to guarantee a connection to an existing database
		panic(err)
	}

	return dbName
}

func (rr *Runner) CollectMigrations() (Migrations, error) {
	sqlMs, err := fs.Glob(rr.Dir, "*.sql")
	if err != nil {
		return nil, err
	}

	migrations := make(Migrations, len(sqlMs))
	errgp := new(errgroup.Group)
	for idx, fname := range sqlMs {
		errgp.Go(func() error {
			sqlfile, err := rr.Dir.Open(fname)
			if err != nil {
				return err
			}
			migrations.FillVersionAtIndex(idx, sqlfile)
			return nil
		})
	}
	if err = errgp.Wait(); err != nil {
		return nil, err
	}

	return migrations, nil
}
