CREATE TABLE users (
    id UInt64,
    username String NOT NULL,
    email String NOT NULL,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
PRIMARY KEY id;

/* rollback
DROP TABLE users;
*/