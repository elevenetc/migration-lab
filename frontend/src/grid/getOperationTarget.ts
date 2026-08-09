import type {Operation} from '../api/migrationApi.ts';

/** What the operation touches: the counterpart of `getOperationTitle`, which says what it does. */
export function getOperationTarget(operation: Operation): string {
    switch (operation.type) {
        case 'CREATE_TABLE':
        case 'DROP_TABLE':
            return operation.tableName;
        case 'ADD_COLUMN':
            return operation.column.name;
        case 'ALTER_COLUMN_TYPE':
        case 'SET_NOT_NULL':
        case 'DROP_NOT_NULL':
        case 'SET_DEFAULT':
        case 'DROP_DEFAULT':
        case 'DROP_COLUMN':
            return operation.columnName;
        case 'RENAME_COLUMN':
            return operation.newColumnName;
        case 'RENAME_TABLE':
            return operation.newTableName;
        case 'ADD_CONSTRAINT':
        case 'DROP_CONSTRAINT':
            return operation.constraintName;
        default: {
            // Never reached: every variant above is handled, so a new one breaks the build here.
            return (operation as Operation).type;
        }
    }
}
