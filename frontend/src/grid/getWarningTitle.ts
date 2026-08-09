import {Warning} from '../api/migrationApi.ts'

export function getWarningTitle(warning: Warning): string {
    if (warning.type == 'ACCESS_EXCLUSIVE_LOCK') {
        return 'exclusive lock'
    } else {
        return (warning as Warning).type
    }
}
