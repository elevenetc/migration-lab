package com.migrationtimeline.models

/**
 * Returns true if:
 * - [a] contains creation of table T
 * - [b] contains operation of altering table T
 */
fun isConnectedWithAsAlter(a: Migration, b: Migration): Boolean {
    val create = a.operations.filterIsInstance<CreateTable>()
    val alter = b.operations.filterIsInstance<AlterTable>()

    alter.forEach { alterTable ->
        create.forEach { createTable ->
            if (createTable.tableName == alterTable.tableName) {
                return true
            }
        }
    }

    return false
}