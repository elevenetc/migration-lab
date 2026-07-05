package com.migrationtimeline.parser.alter

import com.migrationtimeline.models.RenameTable
import net.sf.jsqlparser.statement.alter.Alter
import net.sf.jsqlparser.statement.alter.AlterOperation

fun parseRenameTable(migrationId: String, tableName: String, alter: Alter): RenameTable? {
    return alter.alterExpressions
        ?.firstOrNull { it.operation == AlterOperation.RENAME_TABLE }
        ?.let { expr ->
            val newName = expr.newTableName ?: return@let null
            RenameTable(
                migrationId = migrationId,
                tableName = tableName,
                newTableName = newName
            )
        }
}
