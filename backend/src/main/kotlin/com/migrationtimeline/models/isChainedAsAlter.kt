package com.migrationtimeline.models

/**
 * Returns true if:
 * - first migration is table T creation
 * - all consecutive are alters of the table T (AddColumn operations)
 */
fun isChainedAsAlter(timeline: List<Migration>): Boolean {
    if (timeline.isEmpty()) return false

    val first = timeline.first()
    val createTable = first.operations.filterIsInstance<CreateTable>().firstOrNull() ?: return false
    val tableName = createTable.tableName

    return timeline.drop(1).all { migration ->
        migration.operations.any { op ->
            op is AddColumn && op.tableName == tableName
        }
    }
}