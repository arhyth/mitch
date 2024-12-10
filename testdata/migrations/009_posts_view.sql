CREATE MATERIALIZED VIEW posts_view
ENGINE = MergeTree()
ORDER BY id AS
SELECT
    p.id,
    p.title,
    p.content,
    p.created_at,
    u.username AS author
FROM posts p
INNER JOIN users u ON p.author_id = u.id;

/* rollback
DROP VIEW posts_view;
*/