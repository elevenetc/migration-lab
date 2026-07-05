package com.migrationtimeline.parser

import com.migrationtimeline.models.CreateTable
import com.migrationtimeline.models.Migration
import com.migrationtimeline.models.Operation
import com.migrationtimeline.parser.alter.parseAlter
import net.sf.jsqlparser.parser.CCJSqlParserUtil
import net.sf.jsqlparser.statement.alter.Alter
import net.sf.jsqlparser.statement.drop.Drop
import net.sf.jsqlparser.statement.create.table.CreateTable as JsqlCreateTable

fun parseMigration(migration: Map<String, String>): List<Migration> {
    return migration.entries.mapIndexed { index, entry ->
        parseMigration(entry.key, entry.value, index.toLong())
    }
}

@Suppress("DEPRECATION")
fun parseMigration(id: String, sql: String, timestamp: Long = 0L): Migration {
    val operations = mutableListOf<Operation>()

    // Check for PARTITION OF statements (JSQLParser doesn't support this PostgreSQL syntax)
    val partitionOfMatch = PARTITION_OF_REGEX.find(sql)
    if (partitionOfMatch != null) {
        val tableName = partitionOfMatch.groupValues[1]
        val parentTable = partitionOfMatch.groupValues[2]
        operations.add(
            CreateTable(
                migrationId = id,
                tableName = tableName,
                columns = emptyList(),
                isPartitioned = false,
                partitionOf = parentTable
            )
        )
        return Migration(id = id, version = id, timestamp = timestamp, operations = operations)
    }

    // Parse with JSQLParser (handles PARTITION BY via tableOptionsStrings)
    try {
        val statements = CCJSqlParserUtil.parseStatements(sql).statements
        val parsedOps = statements.flatMap { statement ->
            when (statement) {
                is JsqlCreateTable -> listOf(parseCreateTable(id, statement))
                is Alter -> parseAlter(id, statement)
                is Drop -> parseDropTable(id, statement)?.let { listOf(it) } ?: emptyList()
                else -> emptyList()
            }
        }
        operations.addAll(parsedOps)
    } catch (_: Exception) {
        // JSQLParser failed, try regex fallback for PARTITION BY RANGE
        val partitionByRangeMatch = PARTITION_BY_RANGE_REGEX.find(sql)
        if (partitionByRangeMatch != null) {
            val tableName = partitionByRangeMatch.groupValues[1]
            val columnsSql = partitionByRangeMatch.groupValues[2]
            operations.add(
                CreateTable(
                    migrationId = id,
                    tableName = tableName,
                    columns = parseColumnsFromSql(columnsSql),
                    isPartitioned = true
                )
            )
        }
    }

    return Migration(id = id, version = id, timestamp = timestamp, operations = operations)
}

@Deprecated("Use parseMigration() functions directly", ReplaceWith("parseMigration"))
object MigrationParser {
    fun parse(migration: Map<String, String>) = parseMigration(migration)
    fun parse(id: String, sql: String, timestamp: Long = 0L) = parseMigration(id, sql, timestamp)
}
