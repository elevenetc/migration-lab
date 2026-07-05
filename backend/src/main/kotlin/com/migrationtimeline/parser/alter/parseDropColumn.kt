package com.migrationtimeline.parser.alter

import com.migrationtimeline.models.DropColumn
import net.sf.jsqlparser.statement.alter.Alter
import net.sf.jsqlparser.statement.alter.AlterOperation

fun parseDropColumn(migrationId: String, tableName: String, alter: Alter): List<DropColumn> {
    return alter.alterExpressions
        ?.filter { it.operation == AlterOperation.DROP && it.columnName != null && it.constraintName == null }
        ?.map { expr ->
            DropColumn(
                migrationId = migrationId,
                tableName = tableName,
                columnName = expr.columnName
            )
        } ?: emptyList()
}
