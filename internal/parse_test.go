package internal_test

import (
	"os"
	"strings"
	"testing"

	"github.com/arhyth/mitch/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	sampleUpSQL = "CREATE TABLE IF NOT EXISTS test_table (\n" +
		"    `TenantId` UInt8,\n" +
		"    `AccountId` UInt16,\n" +
		"    `SiteId` UInt32,\n" +
		"    `Time` DateTime,\n" +
		"    `Created` DateTime DEFAULT NOW()\n" +
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
		as.Equal(sampleUpSQL, ver.Up.Statements[0])
		as.NotEmpty(ver.Down.Statements)
		as.Equal(sampleDownSQL, ver.Down.Statements[0])
	})

	t.Run("missing rollback", func(tt *testing.T) {
		reqrd := require.New(tt)
		as := assert.New(tt)
		file, err := os.Open("../testdata/migrations/002_add_new_field_norollback.sql")
		reqrd.Nil(err)
		ver, err := internal.ParseMigration(file)
		reqrd.Nil(err)
		as.NotEmpty(ver.Up.Statements)
		as.Equal(anotherUpSQL, ver.Up.Statements[0])
		as.Empty(ver.Down.Statements)
	})

	t.Run("multi statement", func(tt *testing.T) {
		var multi = `INSERT INTO users (id, username, email)
VALUES
	(1, 'john_doe', 'john@example.com'),
	(2, 'jane_smith', 'jane@example.com'),
	(3, 'alice_wonderland', 'alice@example.com');

INSERT INTO posts (id, title, content, author_id)
VALUES
	(1, 'Introduction to SQL', 'SQL is a powerful language for managing databases...', 1),
	(2, 'Data Modeling Techniques', 'Choosing the right data model is crucial...', 2),
	(3, 'Advanced Query Optimization', 'Optimizing queries can greatly improve...', 1);

INSERT INTO comments (id, post_id, user_id, content)
VALUES
	(1, 1, 3, 'Great introduction! Looking forward to more.'),
	(2, 1, 2, 'SQL can be a bit tricky at first, but practice helps.'),
	(3, 2, 1, 'You covered normalization really well in this post.');

/* rollback
TRUNCATE TABLE comments;
TRUNCATE TABLE posts;
TRUNCATE TABLE users;
*/`

		expectedFirstUp := `INSERT INTO users (id, username, email)
VALUES
	(1, 'john_doe', 'john@example.com'),
	(2, 'jane_smith', 'jane@example.com'),
	(3, 'alice_wonderland', 'alice@example.com');`
		expectedSecondUp := `INSERT INTO posts (id, title, content, author_id)
VALUES
	(1, 'Introduction to SQL', 'SQL is a powerful language for managing databases...', 1),
	(2, 'Data Modeling Techniques', 'Choosing the right data model is crucial...', 2),
	(3, 'Advanced Query Optimization', 'Optimizing queries can greatly improve...', 1);`

		reqrd := require.New(tt)
		as := assert.New(tt)
		rdr := strings.NewReader(multi)
		ver, err := internal.ParseMigration(rdr)
		reqrd.Nil(err)
		reqrd.NotEmpty(ver.Up.Statements)
		as.Len(ver.Up.Statements, 3)
		as.Equal(expectedFirstUp, ver.Up.Statements[0])
		as.Equal(expectedSecondUp, ver.Up.Statements[1])
		reqrd.NotEmpty(ver.Down.Statements)
		as.Len(ver.Down.Statements, 3)
		as.Equal("TRUNCATE TABLE users;", ver.Down.Statements[2])
	})
}
