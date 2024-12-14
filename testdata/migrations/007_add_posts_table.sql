CREATE TABLE posts (
    id UInt64,
    title String NOT NULL,
    content String NOT NULL,
    author_id UInt64 NOT NULL,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
PRIMARY KEY id;

/* rollback
DROP TABLE posts;
*/