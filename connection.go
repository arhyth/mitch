package mitch

import (
	"database/sql"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
)

func Connect(dbURL string) (db *sql.DB, err error) {
	if strings.HasPrefix(dbURL, "clickhouse") {
		db, err = sql.Open("clickhouse", dbURL)
		if err != nil {
			return nil, err
		}
	} else {
		opts, err := GetDBOptions(dbURL)
		if err != nil {
			return nil, err
		}
		db = clickhouse.OpenDB(opts)
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
