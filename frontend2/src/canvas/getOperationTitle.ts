import {Operation} from "../api/migrationApi.ts";

export function getOperationTitle(operation: Operation): string {
    if (operation.type == 'ADD_COLUMN') {
        return '+column'
    } else if (operation.type == 'CREATE_TABLE') {
        if (operation.partitionOf) {
            return `+part of:${operation.partitionOf}: ` + operation.tableName
        }
        const suffix = operation.isPartitioned ? ' [partitioned]' : ''
        return '+table: ' + operation.tableName + suffix
    } else if (operation.type == 'DROP_TABLE') {
        return '-table'
    } else if (operation.type == 'DROP_NOT_NULL') {
        return '-not-null'
    } else if (operation.type == 'DROP_DEFAULT') {
        return '-default'
    } else if (operation.type == 'DROP_CONSTRAINT') {
        return '-constraint'
    } else if (operation.type == 'DROP_COLUMN') {
        return '-column'
    } else if (operation.type == 'RENAME_TABLE') {
        return '~table: ' + operation.newTableName
    } else if (operation.type == 'RENAME_COLUMN') {
        return '~column'
    } else if (operation.type == 'ALTER_COLUMN_TYPE') {
        return '~column-type'
    } else if (operation.type == 'ADD_CONSTRAINT') {
        return '+constraint'
    } else if (operation.type == 'SET_NOT_NULL') {
        return '>not-null'
    } else if (operation.type == 'SET_DEFAULT') {
        return '>default'
    } else {
        // All known types are handled above; fall back for future operation types
        return (operation as Operation).type
    }
}
