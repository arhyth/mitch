package internal_test

import (
	"os"
	"testing"

	"github.com/arhyth/mitch/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	sampleUpSQL = "CREATE TABLE IF NOT EXISTS test_table (\n" +
		"`TenantId` UInt8,\n" +
		"`AccountId` UInt16,\n" +
		"`SiteId` UInt32,\n" +
		"`Time` DateTime,\n" +
		"`Created` DateTime DEFAULT NOW()\n" +
		")\n" +
		"ENGINE = MergeTree\n" +
		"PRIMARY KEY (toStartOfHour(`Time`), TenantId, AccountId, SiteId)\n" +
		"ORDER BY (toStartOfHour(`Time`), TenantId, AccountId, SiteId)\n" +
		"SETTINGS index_granularity = 8192;"

	sampleDownSQL = "DROP TABLE IF EXISTS test_table;"

	anotherUpSQL = "ALTER TABLE test_table ADD COLUMN NewField UInt32;"
)

func TestParseMigration(t *testing.T) {
	t.Run("happy path", func(tt *testing.T) {
		reqrd := require.New(tt)
		as := assert.New(tt)
		file, err := os.Open("../testdata/migrations/001_default_database.sql")
		reqrd.Nil(err)
		ver, err := internal.ParseMigration(file)
		reqrd.Nil(err)
		as.NotEmpty(ver.Up.Statements)
		as.Equal(ver.Up.Statements, sampleUpSQL)
		as.NotEmpty(ver.Down.Statements)
		as.Equal(ver.Down.Statements, sampleDownSQL)
	})

	t.Run("missing rollback", func(tt *testing.T) {
		reqrd := require.New(tt)
		as := assert.New(tt)
		file, err := os.Open("../testdata/migrations/002_add_new_field_norollback.sql")
		reqrd.Nil(err)
		ver, err := internal.ParseMigration(file)
		reqrd.Nil(err)
		as.NotEmpty(ver.Up.Statements)
		as.Equal(ver.Up.Statements, anotherUpSQL)
		as.Empty(ver.Down.Statements)
	})
}
