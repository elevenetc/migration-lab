package com.migrationtimeline.parser.alter

import com.migrationtimeline.models.RenameColumn
import net.sf.jsqlparser.statement.alter.Alter
import net.sf.jsqlparser.statement.alter.AlterOperation

fun parseRenameColumn(migrationId: String, tableName: String, alter: Alter): List<RenameColumn> {
    return alter.alterExpressions
        ?.filter { it.operation == AlterOperation.RENAME && it.columnName != null }
        ?.mapNotNull { expr ->
            val oldName = expr.columnName ?: return@mapNotNull null
            val newName = expr.columnOldName ?: return@mapNotNull null
            RenameColumn(
                migrationId = migrationId,
                tableName = tableName,
                columnName = newName,
                newColumnName = oldName
            )
        } ?: emptyList()
}
