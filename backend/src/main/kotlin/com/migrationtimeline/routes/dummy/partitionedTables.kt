package com.migrationtimeline.routes.dummy

private val createEvents = """
    CREATE TABLE events (
        id SERIAL,
        name VARCHAR(255) NOT NULL,
        year INT NOT NULL
    ) PARTITION BY LIST (year);
""".trimIndent()

private val createEvents2025 = """
    CREATE TABLE events_2025 PARTITION OF events
        FOR VALUES IN (2025);
""".trimIndent()

private val createEvents2026 = """
    CREATE TABLE events_2026 PARTITION OF events
        FOR VALUES IN (2026);
""".trimIndent()

val partitionedTablesMigrations: LinkedHashMap<String, String> = linkedMapOf(
    "V1__create_events" to createEvents,
    "V2__create_events_2025" to createEvents2025,
    "V3__create_events_2026" to createEvents2026
)
