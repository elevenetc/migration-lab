package com.migrationtimeline.routes.dummy

import com.migrationtimeline.models.MigrationTimelineResponse
import com.migrationtimeline.routes.migrationsToResponse

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

private val alterEvents = """
    ALTER TABLE events ADD COLUMN description TEXT;
""".trimIndent()

fun partitionedTableWithAlter(): MigrationTimelineResponse {
    return migrationsToResponse(
        mapOf(
            "V1__create_events" to createEvents,
            "V2__create_events_2025" to createEvents2025,
            "V3__create_events_2026" to createEvents2026,
            "V4__add_description" to alterEvents
        )
    )
}
