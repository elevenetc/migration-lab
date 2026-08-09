import type {Operation} from '../api/migrationApi.ts';
import type {MigrationKind} from './drawers/migration';

/** What the migration does to the table, in the order the kinds outrank each other. */
export function getMigrationKind(operations: Operation[]): MigrationKind {
    if (operations.some(operation => operation.type === 'CREATE_TABLE' && operation.partitionOf)) return 'partition';
    if (operations.some(operation => operation.type === 'CREATE_TABLE')) return 'create';
    if (operations.some(operation => operation.type === 'DROP_TABLE')) return 'drop';
    return 'alter';
}
