package mitch

import (
	"fmt"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/rs/zerolog/log"
)

var (
	EnvMigrationDir     = "CLICKHOUSE_MIGRATION_DIR"
	EnvDatabaseHost     = "CLICKHOUSE_HOST"
	EnvDatabasePort     = "CLICKHOUSE_PORT"
	EnvDatabaseName     = "CLICKHOUSE_DB"
	EnvDatabaseUser     = "CLICKHOUSE_USER"
	EnvDatabasePassword = "CLICKHOUSE_PASSWORD"

	DefaultMigrationDir = "migrations"
	DefaultPort         = "9000"
	// these defaults are of Clickhouse and not ours
	DefaultUser     = "default"
	DefaultPassword = ""
)

func GetDBOptions(dbURL string) (*clickhouse.Options, error) {
	if dbURL == "" {
		if os.Getenv(EnvDatabaseHost) == "" {
			return nil, ErrUnsetDBURL
		}
		port := os.Getenv(EnvDatabasePort)
		if port == "" {
			port = DefaultPort
		}
		dbURL = fmt.Sprintf("tcp://%s:%s", EnvDatabaseHost, port)
	}

	dbName := os.Getenv(EnvDatabaseName)
	if dbName == "" {
		return nil, ErrUnsetDBName
	}
	dbUser := os.Getenv(EnvDatabaseUser)
	if dbUser == "" {
		log.Warn().Msgf("%s not set; using user `default`", EnvDatabaseUser)
		dbUser = DefaultUser
	}
	dbPasswd := os.Getenv(EnvDatabasePassword)
	if dbPasswd == "" {
		if dbUser != DefaultUser {
			return nil, ErrUnsetDBPassword
		}
		dbPasswd = DefaultPassword
	}
	return &clickhouse.Options{
		Addr: []string{dbURL},
		Auth: clickhouse.Auth{
			Database: EnvDatabaseName,
			Username: dbUser,
			Password: dbPasswd,
		},
	}, nil
}
