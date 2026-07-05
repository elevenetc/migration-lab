import {useCallback, useEffect, useMemo} from 'react'
import {Background, BaseEdge, Controls, Edge, EdgeProps, MarkerType, Node, Position, ReactFlow, useEdgesState, useNodesState,} from '@xyflow/react'
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

interface PartitionInfo {
    childTable: string
    parentTable: string
    migrationId: string
}

function formatOperationSummary(operation: Operation): string {
    if (operation.type === 'CREATE_TABLE') {
        if (operation.partitionOf) {
            return `partition(of:${operation.partitionOf})`
        }
        const columnNames = operation.columns.map(c => c.name).join(', ')
        const suffix = operation.isPartitioned ? ' [partitioned]' : ''
        return `create(${columnNames})${suffix}`
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
    if (operation.type === 'DROP_COLUMN') {
        return `dropColumn(${operation.columnName})`
    }
    if (operation.type === 'ADD_COLUMN') {
        return `add(${operation.column.name})`
    }
    console.warn('Unsupported operation type:', operation)
    return 'unknown'
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
    if (operation.type === 'DROP_COLUMN') {
        return '#c05621' // orange - column deletion
    }
    if (operation.type === 'ADD_COLUMN') {
        return '#276749' // green - add column
    }
    return '#718096' // gray - unknown
}

interface TableOperationsResult {
    tableOperations: Map<string, OperationEntry[]>
    renames: RenameInfo[]
    partitions: PartitionInfo[]
    tableOrder: string[] // Explicit row ordering
}

function buildTableOperationsMap(migrations: Migration[]): TableOperationsResult {
    const tableOperations = new Map<string, OperationEntry[]>()
    const renames: RenameInfo[] = []
    const partitions: PartitionInfo[] = []
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
                // Insert partition after parent and its existing children
                } else if (operation.type === 'CREATE_TABLE' && operation.partitionOf) {
                    const parentIndex = tableOrder.indexOf(operation.partitionOf)
                    if (parentIndex !== -1) {
                        // Find last existing child of this parent
                        let insertIndex = parentIndex + 1
                        for (let i = parentIndex + 1; i < tableOrder.length; i++) {
                            const existingOps = tableOperations.get(tableOrder[i])
                            const isChildOfSameParent = existingOps?.some(
                                e => e.operation.type === 'CREATE_TABLE' &&
                                     e.operation.partitionOf === operation.partitionOf
                            )
                            if (isChildOfSameParent) {
                                insertIndex = i + 1
                            } else {
                                break
                            }
                        }
                        tableOrder.splice(insertIndex, 0, groupKey)
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

            // Track partitions for cross-table edge building
            if (operation.type === 'CREATE_TABLE' && operation.partitionOf) {
                partitions.push({
                    childTable: operation.tableName,
                    parentTable: operation.partitionOf,
                    migrationId: migration.id
                })
            }
        })
    })

    return {tableOperations, renames, partitions, tableOrder}
}

const NODE_WIDTH = 140
const NODE_HEIGHT = 60
const X_GAP = 20
const Y_GAP = 15

function StepDownEdge({sourceX, sourceY, targetX, targetY, markerEnd, style}: EdgeProps) {
    // Start from right side of source, go down, then right to target
    const path = `M ${sourceX} ${sourceY} L ${sourceX} ${targetY} L ${targetX} ${targetY}`
    return <BaseEdge path={path} markerEnd={markerEnd} style={style}/>
}

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
                                <div style={{fontWeight: 'bold', fontSize: '9px'}}>{tableName}</div>
                                <div style={{fontSize: '8px', color: '#a0aec0'}}>
                                    {item.migration.version}
                                </div>
                                <div style={{fontSize: '8px', color: '#cbd5e0'}}>
                                    {formatOperationSummary(item.operation)}
                                </div>
                            </div>
                        ),
                    },
                    style: {
                        background: getNodeColor(item.operation),
                        color: 'white',
                        padding: 4,
                        borderRadius: 4,
                        width: NODE_WIDTH,
                        fontSize: '9px',
                    },
                })
            })
        })

        return nodes
    }, [migrations])

    const buildEdges = useCallback((): Edge[] => {
        const {tableOperations, renames, partitions} = buildTableOperationsMap(migrations)
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
                    type: 'stepDown',
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
                    type: 'stepDown',
                    markerEnd: {type: MarkerType.ArrowClosed},
                    style: {strokeDasharray: '5,5'}, // Dashed line to distinguish cross-table edges
                })
            }
        })

        // Build cross-table edges for partitions (from parent CREATE_TABLE to child partition)
        partitions.forEach(partition => {
            const parentOps = tableOperations.get(partition.parentTable)
            if (!parentOps || parentOps.length === 0) {
                console.warn(`Partition ${partition.childTable}: parent table "${partition.parentTable}" not found`)
                return
            }
            const parentCreateOp = parentOps.find(op => op.operation.type === 'CREATE_TABLE')
            if (!parentCreateOp) {
                console.warn(`Partition ${partition.childTable}: parent table "${partition.parentTable}" has no CREATE_TABLE operation`)
                return
            }
            const sourceId = `${parentCreateOp.migration.id}-${partition.parentTable}`
            const targetId = `${partition.migrationId}-${partition.childTable}`
            edges.push({
                id: `partition-${sourceId}->${targetId}`,
                source: sourceId,
                target: targetId,
                type: 'stepDown',
                markerEnd: {type: MarkerType.ArrowClosed},
                style: {strokeDasharray: '3,3', stroke: '#38a169'}, // Dotted green line for partitions
            })
        })

        return edges
    }, [migrations])

    const [nodes, setNodes] = useNodesState<Node>([])
    const [edges, setEdges] = useEdgesState<Edge>([])
    const edgeTypes = useMemo(() => ({stepDown: StepDownEdge}), [])

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
                edgeTypes={edgeTypes}
                fitView
                minZoom={0.1}
                attributionPosition="bottom-left"
            >
                <Background/>
                <Controls/>
            </ReactFlow>
        </div>
    )
}
