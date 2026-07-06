package com.migrationtimeline.analysis

import com.migrationtimeline.models.*

/**
 * Detects ALTER operations on partitioned tables that acquire ACCESS EXCLUSIVE lock
 * on the parent and all partitions.
 *
 * Returns [AccessExclusiveLock] warning for: ADD COLUMN, DROP COLUMN, ALTER COLUMN TYPE,
 * SET/DROP NOT NULL, SET/DROP DEFAULT, ADD/DROP CONSTRAINT, RENAME COLUMN.
 */
internal fun detectAccessExclusiveLock(op: Operation, context: AnalysisContext): Warning? {
    val (tableName, migrationId) = when (op) {
        is AddColumn -> op.tableName to op.migrationId
        is DropColumn -> op.tableName to op.migrationId
        is AlterColumnType -> op.tableName to op.migrationId
        is SetNotNull -> op.tableName to op.migrationId
        is DropNotNull -> op.tableName to op.migrationId
        is SetDefault -> op.tableName to op.migrationId
        is DropDefault -> op.tableName to op.migrationId
        is AddConstraint -> op.tableName to op.migrationId
        is DropConstraint -> op.tableName to op.migrationId
        is RenameColumn -> op.tableName to op.migrationId
        is CreateTable, is RenameTable, is DropTable -> return null
    }

    return if (tableName in context.partitionedTables) {
        AccessExclusiveLock(
            operationId = OperationId(migrationId, tableName),
            tableName = tableName,
            message = "ALTER on partitioned table '$tableName' acquires ACCESS EXCLUSIVE lock on parent and all partitions"
        )
    } else null
}
