package com.migrationtimeline.parser.alter

import com.migrationtimeline.models.Operation
import net.sf.jsqlparser.statement.alter.Alter

fun parseAlter(migrationId: String, alter: Alter): List<Operation> {
    val tableName = alter.table.name
    val operations = mutableListOf<Operation>()

    operations.addAll(parseAddColumn(migrationId, tableName, alter))
    operations.addAll(parseDropColumn(migrationId, tableName, alter))
    operations.addAll(parseAlterColumnType(migrationId, tableName, alter))
    operations.addAll(parseSetNotNull(migrationId, tableName, alter))
    operations.addAll(parseDropNotNull(migrationId, tableName, alter))
    operations.addAll(parseSetDefault(migrationId, tableName, alter))
    operations.addAll(parseDropDefault(migrationId, tableName, alter))
    parseRenameTable(migrationId, tableName, alter)?.let { operations.add(it) }
    operations.addAll(parseRenameColumn(migrationId, tableName, alter))
    operations.addAll(parseAddConstraint(migrationId, tableName, alter))
    operations.addAll(parseDropConstraint(migrationId, tableName, alter))

    return operations
}
