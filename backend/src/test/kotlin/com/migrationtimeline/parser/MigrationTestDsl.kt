package com.migrationtimeline.parser

import com.migrationtimeline.models.AddColumn
import com.migrationtimeline.models.AddConstraint
import com.migrationtimeline.models.AlterColumnType
import com.migrationtimeline.models.CreateTable
import com.migrationtimeline.models.DropColumn
import com.migrationtimeline.models.DropConstraint
import com.migrationtimeline.models.DropDefault
import com.migrationtimeline.models.DropNotNull
import com.migrationtimeline.models.DropTable
import com.migrationtimeline.models.Migration
import com.migrationtimeline.models.Operation
import com.migrationtimeline.models.RenameColumn
import com.migrationtimeline.models.RenameTable
import com.migrationtimeline.models.SetDefault
import com.migrationtimeline.models.SetNotNull
import java.util.UUID
import kotlin.test.assertEquals

fun String.toMigration(): Migration =
    MigrationParser.parse(UUID.randomUUID().toString(), this)

fun List<String>.toMigrations(): List<Migration> =
    map { it.toMigration() }

fun Map<Int, String>.toMigrations(): List<Migration> =
    entries.map { (timestamp, sql) ->
        MigrationParser.parse(timestamp.toString(), sql, timestamp.toLong())
    }

fun List<Pair<Int, String>>.toTimedMigrations(): List<Migration> =
    mapIndexed { index, (timestamp, sql) ->
        MigrationParser.parse("$timestamp-$index", sql, timestamp.toLong())
    }

fun Migration.isEqualTo(expectedDsl: String) {
    assertEquals(expectedDsl, toTestDsl())
}

fun List<Migration>.isEqualTo(expectedDsl: String) {
    assertEquals(expectedDsl, toTestDsl())
}

fun List<Migration>.isEqualToTimed(expectedDsl: String) {
    assertEquals(expectedDsl, toTimedTestDsl())
}

fun List<Migration>.toTestDsl(): String =
    flatMap { it.operations }
        .groupBy { it.tableName() }
        .toSortedMap()
        .map { (_, ops) -> ops.joinToString(" > ") { it.toTestDsl() } }
        .joinToString("\n")

fun List<Migration>.toTimedTestDsl(): String {
    data class OpWithTime(val op: Operation, val timestamp: Long)

    val opsWithTime = flatMap { migration ->
        migration.operations.map { OpWithTime(it, migration.timestamp) }
    }

    val grouped = opsWithTime.groupBy { it.op.tableName() }

    // Sort tables by their first operation's timestamp
    val sortedTables = grouped.entries.sortedBy { (_, ops) -> ops.minOf { it.timestamp } }

    val baseTimestamp = sortedTables.firstOrNull()?.value?.minOf { it.timestamp } ?: 0L

    return sortedTables.map { (_, ops) ->
        val firstTimestamp = ops.minOf { it.timestamp }
        val depth = (firstTimestamp - baseTimestamp).toInt()
        val prefix = ">".repeat(depth)
        val opsStr = ops.sortedBy { it.timestamp }.joinToString(" > ") { it.op.toTestDsl() }
        "$prefix$opsStr"
    }.joinToString("\n")
}

fun Migration.toTestDsl(): String =
    operations.joinToString(" > ") { it.toTestDsl() }

fun Operation.tableName(): String = when (this) {
    is CreateTable -> tableName
    is AddColumn -> tableName
    is AlterColumnType -> tableName
    is SetNotNull -> tableName
    is DropNotNull -> tableName
    is SetDefault -> tableName
    is DropDefault -> tableName
    is RenameTable -> tableName
    is RenameColumn -> tableName
    is AddConstraint -> tableName
    is DropConstraint -> tableName
    is DropTable -> tableName
    is DropColumn -> tableName
}

private fun Operation.toTestDsl(): String = when (this) {
    is CreateTable -> when {
        partitionOf != null -> "createPartition(${tableName}:${partitionOf})"
        isPartitioned -> "createPartitioned(${tableName}(${columns.joinToString(",") { it.name }}))"
        else -> "create(${tableName}(${columns.joinToString(",") { it.name }}))"
    }
    is AddColumn -> "addColumn(${tableName}(${column.name}))"
    is AlterColumnType -> "alterType(${tableName}(${columnName}:${newType}))"
    is SetNotNull -> "setNotNull(${tableName}(${columnName}))"
    is DropNotNull -> "dropNotNull(${tableName}(${columnName}))"
    is SetDefault -> "setDefault(${tableName}(${columnName}=${defaultValue}))"
    is DropDefault -> "dropDefault(${tableName}(${columnName}))"
    is RenameTable -> "renameTable(${tableName}->${newTableName})"
    is RenameColumn -> "renameColumn(${tableName}(${columnName}->${newColumnName}))"
    is AddConstraint -> "addConstraint(${tableName}(${constraintName}:${constraintType}))"
    is DropConstraint -> "dropConstraint(${tableName}(${constraintName}))"
    is DropTable -> "drop(${tableName})"
    is DropColumn -> "dropColumn(${tableName}(${columnName}))"
}
