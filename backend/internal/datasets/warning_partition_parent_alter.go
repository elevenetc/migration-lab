package datasets

var warningPartitionParentAlterMigrations = []migration{
	{"v1", createEventsPartitionedByYear},
	{"v2", createEvents2026},
	{"v3", alterEventsAddArchived},
}

const createEventsPartitionedByYear = `CREATE TABLE events (
    id SERIAL,
    year INT NOT NULL
) PARTITION BY RANGE (year);`

const createEvents2026 = `CREATE TABLE events_2026 PARTITION OF events
    FOR VALUES FROM (2026) TO (2027);`

const alterEventsAddArchived = `ALTER TABLE events ADD COLUMN archived BOOLEAN;`
