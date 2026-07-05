package com.migrationtimeline.parser.alter

import com.migrationtimeline.models.DropConstraint
import net.sf.jsqlparser.statement.alter.Alter
import net.sf.jsqlparser.statement.alter.AlterOperation

fun parseDropConstraint(migrationId: String, tableName: String, alter: Alter): List<DropConstraint> {
    return alter.alterExpressions
        ?.filter { it.operation == AlterOperation.DROP && it.constraintName != null }
        ?.map { expr ->
            DropConstraint(
                migrationId = migrationId,
                tableName = tableName,
                constraintName = expr.constraintName
            )
        } ?: emptyList()
}
