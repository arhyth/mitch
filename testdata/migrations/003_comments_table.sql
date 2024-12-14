CREATE TABLE comments (
    id UInt64,
    post_id UInt64 NOT NULL,
    user_id UInt64 NOT NULL,
    content String NOT NULL,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
PRIMARY KEY id;

/* rollback
DROP TABLE comments;
*/