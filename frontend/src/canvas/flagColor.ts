import {Operation} from '../api/migrationApi.ts'
import {getOperationColor} from './getOperationColor.ts'
import {getTableColor} from './getTableColor.ts'

export interface FlagColor {
    from: string  // gradient stop 0 (left)
    to: string    // gradient stop 1 (right)
    text: string  // label color
}

export function operationFlagColor(operation: Operation): FlagColor {
    return {from: getTableColor(operation), to: getOperationColor(operation), text: '#ffffff'}
}

export const warningFlagColor: FlagColor = {from: '#f59e0b', to: '#b45309', text: '#ffffff'}
