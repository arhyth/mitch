package mitch

import (
	"database/sql"
	"io/fs"
)

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
