package com.migrationtimeline.routes

import com.migrationtimeline.models.CreateTable
import com.migrationtimeline.models.MigrationTimelineResponse
import com.migrationtimeline.parser.MigrationParser

fun migrationsToResponse(migrationFiles: Map<String, String>): MigrationTimelineResponse {
    val timeline = MigrationParser.parse(migrationFiles)
    val map = timeline.associateBy { t -> t.id }
    val createTable = timeline.flatMap { t -> t.operations }
        .filterIsInstance<CreateTable>()
        .associateBy { ct -> ct.tableName }


    return MigrationTimelineResponse(
        timeline = timeline,
        map = map,
        createTableMap = createTable
    )
}