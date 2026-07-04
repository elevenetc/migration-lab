package com.migrationtimeline.parser

import com.migrationtimeline.models.AlterTable
import com.migrationtimeline.models.CreateTable
import com.migrationtimeline.models.Migration
import com.migrationtimeline.models.Operation
import java.util.UUID
import kotlin.test.assertEquals

fun String.toMigration(): Migration =
    MigrationParser.parse(UUID.randomUUID().toString(), this)

fun List<String>.toMigrations(): List<Migration> =
    map { it.toMigration() }

fun Migration.isEqualTo(expectedDsl: String) {
    assertEquals(expectedDsl, toTestDsl())
}

fun List<Migration>.isEqualTo(expectedDsl: String) {
    assertEquals(expectedDsl, toTestDsl())
}

fun List<Migration>.toTestDsl(): String =
    flatMap { it.operations }
        .groupBy { it.tableName() }
        .toSortedMap()
        .map { (_, ops) -> ops.joinToString(" > ") { it.toTestDsl() } }
        .joinToString("\n")

fun Migration.toTestDsl(): String =
    operations.joinToString(" > ") { it.toTestDsl() }

private fun Operation.tableName(): String = when (this) {
    is CreateTable -> tableName
    is AlterTable -> tableName
}

private fun Operation.toTestDsl(): String = when (this) {
    is CreateTable -> "create(${tableName}(${columns.joinToString(",") { it.name }}))"
    is AlterTable -> "alter(${tableName}(${addedColumns.joinToString(",") { "+${it.name}" }}))"
}
