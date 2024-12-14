CREATE TABLE IF NOT EXISTS test_table (
    `TenantId` UInt8,
    `AccountId` UInt16,
    `SiteId` UInt32,
    `Time` DateTime,
    `Created` DateTime DEFAULT NOW()
)
ENGINE = MergeTree
PRIMARY KEY (toStartOfHour(`Time`), TenantId, AccountId, SiteId)
ORDER BY (toStartOfHour(`Time`), TenantId, AccountId, SiteId)
SETTINGS index_granularity = 8192;


/* rollback
DROP TABLE IF EXISTS test_table;
*/
