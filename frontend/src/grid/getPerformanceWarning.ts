import type {Operation} from '../api/migrationApi.ts';
import type {CellWarning} from './drawers/migration';
import {getOperationTarget} from './getOperationTarget';
import {getOperationTitle} from './getOperationTitle';

/**
 * The two expensive performance classes read as warnings on the timeline;
 * `METADATA_ONLY` is the unremarkable case and raises nothing.
 */
export function getPerformanceWarning(operation: Operation): CellWarning | null {
    const what = `${getOperationTitle(operation)} ${getOperationTarget(operation)}`;
    switch (operation.performanceClass) {
        case 'DATA_SCANNING':
            return {
                title: 'data scan',
                message: `${what} reads every existing row to validate it or to build an index; cost scales with table size`,
            };
        case 'TABLE_REWRITE':
            return {
                title: 'table rewrite',
                message: `${what} rewrites the whole table on disk; cost scales with table size`,
            };
        default:
            return null;
    }
}
