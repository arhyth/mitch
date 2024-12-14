package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2"
)

const envTemplate = `# ClickHouse Configuration
CLICKHOUSE_HOST=clickhouse-server
CLICKHOUSE_MIGRATION_DIR=/app/testdata/migrations/
CLICKHOUSE_DB={{ .DBName }}
`

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

func createEnvFile(path, dbName string) error {
	// Parse the template
	tmpl, err := template.New("env").Parse(envTemplate)
	if err != nil {
		return err
	}

	// Generate the .env content
	outputFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	err = tmpl.Execute(outputFile, EnvTempl{DBName: dbName})
	return err
}
