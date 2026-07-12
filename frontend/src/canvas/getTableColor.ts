import {Operation} from "../api/migrationApi.ts";
import {getColorFromString} from "./getColorFromString.ts";

// Returns a stable color for the table an operation applies to.
// For RENAME_TABLE the new name is used, so the renamed table keeps a single color.
export function getTableColor(operation: Operation): string {
    const tableName = operation.type == 'RENAME_TABLE'
        ? operation.newTableName
        : operation.tableName
    return getColorFromString(tableName)
}
