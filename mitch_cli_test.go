package mitch_test

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"text/template"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/stretchr/testify/require"
)

const envTemplate = `# ClickHouse Configuration
CLICKHOUSE_HOST=clickhouse-server
CLICKHOUSE_MIGRATION_DIR=/app/testdata/migrations/
CLICKHOUSE_DB={{ .DBName }}
`

func TestBinaryForward(t *testing.T) {
	cli := buildCLI(t)

	t.Run("ok", func(t *testing.T) {
		t.Skip()
		total := countSQLFiles(t, "/app/testdata/migrations")

		expectOut := "mitch: successfully migrated database to version: " + strconv.Itoa(total)
		out, err := cli.run("--config", "/app/testdata/test.env")
		require.NoError(t, err)
		require.Contains(t, out, expectOut)
	})

	t.Run("nonsequential", func(t *testing.T) {
		expectOut := "mitch: successfully migrated database to version: 9"
		out, err := cli.run("--config", "/app/testdata/nonseq.env")
		require.NoError(t, err)
		require.Contains(t, out, expectOut)
	})
}

func TestBinaryRollback(t *testing.T) {
	t.Skip()
	cli := buildCLI(t)

	t.Run("ok", func(t *testing.T) {
		reqrd := require.New(t)
		_, err := cli.run("--config", "/app/testdata/test.env")
		reqrd.NoError(err)
		out, err := cli.run(
			"--config",
			"/app/testdata/test.env",
			"--rollback",
			"001_default_database.sql",
		)
		expectOut := "mitch: successfully rolled database back to version: 0"
		reqrd.NoError(err)
		reqrd.Contains(out, expectOut)
	})
}

type testenv struct {
	binaryPath string
	envPath    string
}

func (e testenv) run(params ...string) (string, error) {
	cmd := exec.Command(e.binaryPath, params...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run mitch: %v\nout: %v", err, string(out))
	}
	return string(out), nil
}

func buildCLI(t *testing.T) testenv {
	t.Helper()
	binName := "mitch"
	dir := t.TempDir()
	binOut := filepath.Join(dir, binName)
	args := []string{
		"build",
		"-o", binOut,
		"./cmd/mitch",
	}

	build := exec.Command("go", args...)
	out, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build %s test binary: %v: %s", binName, err, string(out))
	}

	return testenv{
		binaryPath: binOut,
		envPath:    filepath.Join(dir, "test.env"),
	}
}

func countSQLFiles(t *testing.T, dir string) int {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	require.NoError(t, err)
	return len(files)
}

type EnvTempl struct {
	DBName string
}

// dbHelper wraps *sql.DB and provides helper methods
type dbHelper struct {
	db *sql.DB
}

// NewDBHelper initializes a new dbHelper
func NewDBHelper() (*dbHelper, error) {
	opts := &clickhouse.Options{
		Addr: []string{"clickhouse-server:9000"},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
		},
	}
	db := clickhouse.OpenDB(opts)
	return &dbHelper{db: db}, nil
}

// CreateDatabase creates the database specified in the config
func (d *dbHelper) CreateDatabase(dbName string) error {
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", dbName)
	_, err := d.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

// DropDatabase drops the database specified in the config
func (d *dbHelper) DropDatabase(dbName string) error {
	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s;", dbName)
	_, err := d.db.Exec(query)
	return err
}

func createEnvFile(dbName string) {
	// Parse the template
	tmpl, err := template.New("env").Parse(envTemplate)
	if err != nil {
		panic(err)
	}

	// Generate the .env content
	outputFile, err := os.Create("test.env")
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	if err := tmpl.Execute(outputFile, EnvTempl{DBName: dbName}); err != nil {
		panic(err)
	}
}
