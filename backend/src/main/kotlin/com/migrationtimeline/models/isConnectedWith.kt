package com.migrationtimeline.models

/**
 * Returns true if:
 * - [a] contains creation of table T
 * - [b] contains operation of adding column to table T
 */
fun isConnectedWithAsAlter(a: Migration, b: Migration): Boolean {
    val create = a.operations.filterIsInstance<CreateTable>()
    val addColumns = b.operations.filterIsInstance<AddColumn>()

    addColumns.forEach { addColumn ->
        create.forEach { createTable ->
            if (createTable.tableName == addColumn.tableName) {
                return true
            }
        }
    }

    return false
}