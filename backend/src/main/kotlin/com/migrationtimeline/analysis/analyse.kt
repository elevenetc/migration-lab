package com.migrationtimeline.analysis

import com.migrationtimeline.models.*

internal data class AnalysisContext(
    val partitionedTables: Set<String>
)

fun analyse(migrations: List<Migration>): AnalysisResult {
    val partitionedTables = migrations
        .flatMap { it.operations }
        .filterIsInstance<CreateTable>()
        .filter { it.isPartitioned }
        .map { it.tableName }
        .toSet()

    val context = AnalysisContext(partitionedTables)

    val warnings = migrations.flatMap { migration ->
        migration.operations.flatMap { analyzeOperation(it, context) }
    }

    return AnalysisResult(warnings)
}

private fun analyzeOperation(op: Operation, context: AnalysisContext): List<Warning> = listOfNotNull(
    detectAccessExclusiveLock(op, context)
)
