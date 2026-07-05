import {useCallback, useEffect} from 'react'
import {Background, Controls, Edge, MarkerType, Node, Position, ReactFlow, useEdgesState, useNodesState,} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import {useMigrationStore} from '../store/migrationStore'
import type {Migration, Operation} from '../api/migrationApi'

interface OperationEntry {
    migration: Migration
    operation: Operation
}

interface RenameInfo {
    fromTable: string
    toTable: string
    migrationId: string
}

function formatOperationSummary(operation: Operation): string {
    if (operation.type === 'CREATE_TABLE') {
        const columnNames = operation.columns.map(c => c.name).join(', ')
        return `create(${columnNames})`
    }
    if (operation.type === 'ALTER_COLUMN_TYPE') {
        return `type(${operation.columnName}→${operation.newType})`
    }
    if (operation.type === 'SET_NOT_NULL') {
        return `notNull(${operation.columnName})`
    }
    if (operation.type === 'DROP_NOT_NULL') {
        return `nullable(${operation.columnName})`
    }
    if (operation.type === 'SET_DEFAULT') {
        return `default(${operation.columnName}=${operation.defaultValue})`
    }
    if (operation.type === 'DROP_DEFAULT') {
        return `dropDefault(${operation.columnName})`
    }
    if (operation.type === 'RENAME_TABLE') {
        return `rename(→${operation.newTableName})`
    }
    if (operation.type === 'RENAME_COLUMN') {
        return `rename(${operation.columnName}→${operation.newColumnName})`
    }
    if (operation.type === 'ADD_CONSTRAINT') {
        return `constraint(${operation.constraintName}:${operation.constraintType})`
    }
    if (operation.type === 'DROP_CONSTRAINT') {
        return `dropConstraint(${operation.constraintName})`
    }
    if (operation.type === 'DROP_TABLE') {
        return `drop()`
    }
    const added = operation.addedColumns.map(c => `+${c.name}`)
    const dropped = operation.droppedColumns.map(c => `-${c}`)
    const allChanges = [...added, ...dropped].join(', ')
    return `alter(${allChanges})`
}

function getNodeColor(operation: Operation): string {
    if (operation.type === 'CREATE_TABLE') {
        return '#2c5282' // blue
    }
    if (operation.type === 'ALTER_COLUMN_TYPE') {
        return '#6b46c1' // purple
    }
    if (operation.type === 'SET_NOT_NULL') {
        return '#c53030' // red - constraint addition
    }
    if (operation.type === 'DROP_NOT_NULL') {
        return '#dd6b20' // orange - constraint removal
    }
    if (operation.type === 'SET_DEFAULT') {
        return '#319795' // teal - default value
    }
    if (operation.type === 'DROP_DEFAULT') {
        return '#d69e2e' // yellow - default removal
    }
    if (operation.type === 'RENAME_TABLE') {
        return '#9f7aea' // purple - rename table
    }
    if (operation.type === 'RENAME_COLUMN') {
        return '#b794f4' // lighter purple - rename column
    }
    if (operation.type === 'ADD_CONSTRAINT') {
        return '#805ad5' // purple - constraint operations
    }
    if (operation.type === 'DROP_CONSTRAINT') {
        return '#e53e3e' // red - constraint removal
    }
    if (operation.type === 'DROP_TABLE') {
        return '#742a2a' // dark red - table deletion
    }
    if (operation.droppedColumns.length > 0) {
        return '#c05621' // orange
    }
    return '#276749' // green
}

interface TableOperationsResult {
    tableOperations: Map<string, OperationEntry[]>
    renames: RenameInfo[]
    tableOrder: string[] // Explicit row ordering
}

function buildTableOperationsMap(migrations: Migration[]): TableOperationsResult {
    const tableOperations = new Map<string, OperationEntry[]>()
    const renames: RenameInfo[] = []
    const tableOrder: string[] = []

    migrations.forEach(migration => {
        migration.operations.forEach(operation => {
            // For RENAME_TABLE, key by newTableName so it appears on the new table's row
            const groupKey = operation.type === 'RENAME_TABLE'
                ? operation.newTableName
                : operation.tableName

            if (!tableOperations.has(groupKey)) {
                tableOperations.set(groupKey, [])

                // Insert renamed table immediately after its source table
                if (operation.type === 'RENAME_TABLE') {
                    const sourceIndex = tableOrder.indexOf(operation.tableName)
                    if (sourceIndex !== -1) {
                        tableOrder.splice(sourceIndex + 1, 0, groupKey)
                    } else {
                        tableOrder.push(groupKey)
                    }
                } else {
                    tableOrder.push(groupKey)
                }
            }
            tableOperations.get(groupKey)!.push({migration, operation})

            // Track renames for cross-table edge building
            if (operation.type === 'RENAME_TABLE') {
                renames.push({
                    fromTable: operation.tableName,
                    toTable: operation.newTableName,
                    migrationId: migration.id
                })
            }
        })
    })

    return {tableOperations, renames, tableOrder}
}

const NODE_WIDTH = 180
const NODE_HEIGHT = 80
const X_GAP = 40
const Y_GAP = 30

export function Timeline() {
    const {migrations, loading, error} = useMigrationStore()

    const buildNodes = useCallback((): Node[] => {
        const {tableOperations, tableOrder} = buildTableOperationsMap(migrations)
        const nodes: Node[] = []

        const uniqueTimestamps = [...new Set(migrations.map(m => m.timestamp))].sort((a, b) => a - b)
        const timestampRank = new Map(uniqueTimestamps.map((ts, idx) => [ts, idx]))

        tableOrder.forEach((tableName, rowIndex) => {
            const ops = tableOperations.get(tableName) ?? []
            ops.forEach((item) => {
                // For RENAME_TABLE, use newTableName as part of node ID (matches group key)
                const nodeTableKey = item.operation.type === 'RENAME_TABLE'
                    ? item.operation.newTableName
                    : item.operation.tableName
                const nodeId = `${item.migration.id}-${nodeTableKey}`
                const rank = timestampRank.get(item.migration.timestamp) ?? 0

                nodes.push({
                    id: nodeId,
                    type: 'default',
                    position: {
                        x: rank * (NODE_WIDTH + X_GAP),
                        y: rowIndex * (NODE_HEIGHT + Y_GAP)
                    },
                    sourcePosition: Position.Right,
                    targetPosition: Position.Left,
                    data: {
                        label: (
                            <div>
                                <div style={{fontWeight: 'bold', fontSize: '11px'}}>{tableName}</div>
                                <div style={{fontSize: '10px', color: '#a0aec0'}}>
                                    {item.migration.version}
                                </div>
                                <div style={{fontSize: '10px', color: '#cbd5e0'}}>
                                    {formatOperationSummary(item.operation)}
                                </div>
                            </div>
                        ),
                    },
                    style: {
                        background: getNodeColor(item.operation),
                        color: 'white',
                        padding: 8,
                        borderRadius: 6,
                        width: NODE_WIDTH,
                    },
                })
            })
        })

        return nodes
    }, [migrations])

    const buildEdges = useCallback((): Edge[] => {
        const {tableOperations, renames} = buildTableOperationsMap(migrations)
        const edges: Edge[] = []

        // Build same-table edges
        tableOperations.forEach((ops, tableName) => {
            for (let i = 0; i < ops.length - 1; i++) {
                const sourceId = `${ops[i].migration.id}-${tableName}`
                const targetId = `${ops[i + 1].migration.id}-${tableName}`
                edges.push({
                    id: `${sourceId}->${targetId}`,
                    source: sourceId,
                    target: targetId,
                    type: 'smoothstep',
                    markerEnd: {type: MarkerType.ArrowClosed},
                })
            }
        })

        // Build cross-table edges for renames (diagonal from old table's last op to rename node)
        renames.forEach(rename => {
            const sourceOps = tableOperations.get(rename.fromTable)
            if (sourceOps && sourceOps.length > 0) {
                const lastSourceOp = sourceOps[sourceOps.length - 1]
                const sourceId = `${lastSourceOp.migration.id}-${rename.fromTable}`
                const targetId = `${rename.migrationId}-${rename.toTable}`
                edges.push({
                    id: `rename-${sourceId}->${targetId}`,
                    source: sourceId,
                    target: targetId,
                    type: 'smoothstep',
                    markerEnd: {type: MarkerType.ArrowClosed},
                    style: {strokeDasharray: '5,5'}, // Dashed line to distinguish cross-table edges
                })
            }
        })

        return edges
    }, [migrations])

    const [nodes, setNodes] = useNodesState<Node>([])
    const [edges, setEdges] = useEdgesState<Edge>([])

    useEffect(() => {
        setNodes(buildNodes())
        setEdges(buildEdges())
    }, [migrations, buildNodes, buildEdges, setNodes, setEdges])

    if (loading) {
        return <div style={{padding: 20}}>Loading migrations...</div>
    }

    if (error) {
        return <div style={{padding: 20, color: 'red'}}>Error: {error}</div>
    }

    if (migrations.length === 0) {
        return <div style={{padding: 20}}>No migrations loaded</div>
    }

    return (
        <div style={{width: '100%', height: '100%'}}>
            <ReactFlow
                nodes={nodes}
                edges={edges}
                fitView
                attributionPosition="bottom-left"
            >
                <Background/>
                <Controls/>
            </ReactFlow>
        </div>
    )
}
