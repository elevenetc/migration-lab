package com.migrationtimeline.parser.alter

import com.migrationtimeline.models.AlterColumnType
import net.sf.jsqlparser.statement.alter.Alter
import net.sf.jsqlparser.statement.alter.AlterOperation

fun parseAlterColumnType(migrationId: String, tableName: String, alter: Alter): List<AlterColumnType> {
    return alter.alterExpressions
        ?.filter { it.operation == AlterOperation.ALTER && it.colDataTypeList != null }
        ?.flatMap { expr ->
            expr.colDataTypeList
                .filter { !it.isSetNotNull() && !it.isDropNotNull() && !it.isSetDefault() && !it.isDropDefault() }
                .map { colDataType ->
                    AlterColumnType(
                        migrationId = migrationId,
                        tableName = tableName,
                        columnName = colDataType.columnName,
                        newType = colDataType.colDataType.toString().replace(" ", "")
                    )
                }
        } ?: emptyList()
}
