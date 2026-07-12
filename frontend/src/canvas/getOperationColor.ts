import {Operation} from "../api/migrationApi.ts";

let addColor = '#28b828';
let removeColor = '#ff5900';
let changeColor = '#2682cd';

export function getOperationColor(operation: Operation): string {

    let undefinedOperationColor = '#434343';
    if (operation.type == 'ADD_COLUMN')
        return addColor
    else if (operation.type == 'CREATE_TABLE')
        return addColor
    else if (operation.type == 'ADD_CONSTRAINT')
        return addColor
    else if (operation.type == 'DROP_NOT_NULL')
        return removeColor
    else if (operation.type == 'DROP_CONSTRAINT')
        return removeColor
    else if (operation.type == 'DROP_DEFAULT')
        return removeColor
    else if (operation.type == 'DROP_TABLE')
        return removeColor
    else if (operation.type == 'DROP_COLUMN')
        return removeColor
    else if (operation.type == 'RENAME_TABLE')
        return changeColor
    else if (operation.type == 'RENAME_COLUMN')
        return changeColor
    else if (operation.type == 'ALTER_COLUMN_TYPE')
        return changeColor
    else if (operation.type == 'SET_NOT_NULL')
        return changeColor
    else if (operation.type == 'SET_DEFAULT')
        return changeColor
    else return undefinedOperationColor
}
