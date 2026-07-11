package dummy

var simplePartitionMigrations = []migration{
	{"v1", createEventsPartitioned},
	{"v2", createEvents2024},
	{"v3", createEvents2025},
}

const createEventsPartitioned = `CREATE TABLE events (
    id SERIAL,
    created_at TIMESTAMP NOT NULL
) PARTITION BY RANGE (created_at);`

const createEvents2024 = `CREATE TABLE events_2024 PARTITION OF events
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');`

const createEvents2025 = `CREATE TABLE events_2025 PARTITION OF events
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');`
