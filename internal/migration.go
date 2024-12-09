package internal

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"sync"

	"github.com/arhyth/mitch"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type Version struct {
	ID          int64
	ContentHash string
	Up, Down    *SQL
	Source      string
}

type SQL struct {
	Statements string
}

type Migration []Version

func (ms Migration) FillVersionAtIndex(idx int, dir fs.FS, fname string) error {
	file, err := dir.Open(fname)
	if err != nil {
		return err
	}
	ver, err := ParseMigration(file)
	if err != nil {
		return err
	}
	ver.ID = int64(idx + 1)
	ver.Source = fname
	ms[idx] = *ver
	return nil
}

type Runner struct {
	dbNameFunc func() string
	dir        fs.FS
	db         *sql.DB
}

func NewRunner(dir fs.FS, db *sql.DB) *Runner {
	f := sync.OnceValue(func() string {
		query := "SELECT currentDatabase();"
		var dbName string
		err := db.QueryRow(query).Scan(&dbName)
		if err != nil {
			// panic here since the database connection setup in `Connect`
			// is supposed to guarantee a connection to an existing database
			panic(err)
		}
		return dbName
	})

	return &Runner{
		dbNameFunc: f,
		dir:        dir,
		db:         db,
	}
}

func (rr *Runner) Migrate(ctx context.Context) error {
	foundVersions, err := rr.CollectMigration()
	if err != nil {
		return err
	}

	if err := rr.MustVersionTable(ctx); err != nil {
		return err
	}
	dbVersions, err := rr.ListDBVersions(ctx)
	if err != nil {
		return err
	}

	toApply, hasMissing := FindUnappliedVersions(dbVersions, foundVersions)
	if hasMissing {
		log.Warn().Msg("Migration have missing versions")
	}
	var current int64
	for _, ver := range toApply {
		if err := rr.RunUp(ctx, ver); err != nil {
			return err
		}
		current = ver.ID
	}

	if len(toApply) == 0 {
		ver, err := rr.GetCurrentVersion(ctx)
		if err != nil {
			return err
		}
		current = ver.ID

		log.Info().Msgf("mitch: no migrations to run. current version: %d\n", current)
	} else {
		log.Info().Msgf("mitch: successfully migrated database to version: %d\n", current)
	}

	return nil
}

// TODO
func (rr *Runner) RollbackTo(fname string) error {
	return nil
}

func (rr *Runner) RunUp(ctx context.Context, ver Version) error {
	tx, err := rr.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if _, err := tx.ExecContext(ctx, ver.Up.Statements); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to execute SQL: %w", err)
	}

	if err := rr.InsertVersion(ctx, tx, ver); err != nil {
		log.Info().Msg("Rollback transaction")
		_ = tx.Rollback()
		return fmt.Errorf("failed to insert new version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (rr *Runner) InsertVersion(ctx context.Context, tx *sql.Tx, ver Version) error {
	q := `INSERT INTO %s.%s (version_id, source, content_hash) VALUES ($1, $2, $3)`
	_, err := tx.ExecContext(
		ctx,
		fmt.Sprintf(q, rr.GetDBName(), mitch.VersionTable),
		ver.ID,
		ver.Source,
		ver.ContentHash,
	)
	return err
}

func (rr *Runner) MustVersionTable(ctx context.Context) error {
	var query = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.%s (
		    version_id Int64,
		    source String,
			content_hash FixedString(64),
		    created_at DateTime default now()
		)
		ENGINE = MergeTree()
		ORDER BY (version_id, content_hash);
	`, rr.GetDBName(), mitch.VersionTable)
	_, err := rr.db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	tx, err := rr.db.Begin()
	if err != nil {
		return err
	}

	if err = rr.InsertVersion(ctx, tx, Version{ContentHash: "", ID: 0}); err != nil {
		if rberr := tx.Rollback(); rberr != nil {
			log.Error().
				Err(rberr).
				Msg("failed InserVersion rollback")
		}
		return err
	}

	return nil
}

func (rr *Runner) GetDBName() string {
	return rr.dbNameFunc()
}

func (rr *Runner) CollectMigration() (Migration, error) {
	sqlMs, err := fs.Glob(rr.dir, "*.sql")
	if err != nil {
		return nil, err
	}
	sort.Strings(sqlMs)

	migrations := make(Migration, len(sqlMs))
	errgp := new(errgroup.Group)
	for idx, fname := range sqlMs {
		errgp.Go(func() error {
			migrations.FillVersionAtIndex(idx, rr.dir, fname)
			return nil
		})
	}
	if err = errgp.Wait(); err != nil {
		return nil, err
	}

	return migrations, nil
}

func (rr *Runner) ListDBVersions(ctx context.Context) ([]Version, error) {
	q := `SELECT version_id, content_hash, source FROM %s.%s ORDER BY version_id DESC`

	rows, err := rr.db.QueryContext(ctx, fmt.Sprintf(q, rr.GetDBName(), mitch.VersionTable))
	if err != nil {
		return nil, fmt.Errorf("failed to list migrations: %w", err)
	}
	defer rows.Close()

	var versions []Version
	for rows.Next() {
		var ver Version
		if err := rows.Scan(&ver.ID, &ver.ContentHash, &ver.Source); err != nil {
			return nil, fmt.Errorf("failed to scan list migrations result: %w", err)
		}
		versions = append(versions, ver)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return versions, nil
}

func (rr *Runner) GetCurrentVersion(ctx context.Context) (*Version, error) {
	q := `
		SELECT version_id, content_hash, source
		FROM %s.%s
		ORDER BY version_id DESC
		LIMIT 1;
	`

	row := rr.db.QueryRowContext(ctx, fmt.Sprintf(q, rr.GetDBName(), mitch.VersionTable))

	var ver Version
	if err := row.Scan(&ver.ID, &ver.ContentHash, &ver.Source); err != nil {
		return nil, fmt.Errorf("failed to scan version result: %w", err)
	}

	return &ver, nil
}

// FindUnappliedVersions collects versions in the filesystem that has not been applied to the database.
// It does not include missing versions that are versions lower than current version in the database,
// only indicating missing versions by returning a boolean
func FindUnappliedVersions(dbVersions, fsVersions Migration) (unapplied Migration, hasMissing bool) {
	appliedVers := make(map[string]int64)
	for _, ver := range dbVersions {
		appliedVers[ver.ContentHash] = ver.ID
	}

	var dbLatest int64
	for _, dv := range dbVersions {
		if dv.ID > dbLatest {
			dbLatest = dv.ID
		}
	}

	for _, found := range fsVersions {
		_, applied := appliedVers[found.ContentHash]
		if applied {
			continue
		}
		if !applied && found.ID < dbLatest {
			hasMissing = true
			continue
		}

		unapplied = append(unapplied, found)
	}
	sort.SliceStable(unapplied, func(i, j int) bool {
		return unapplied[i].ID < unapplied[j].ID
	})

	return unapplied, hasMissing
}
