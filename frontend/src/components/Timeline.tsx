import {useCallback, useEffect, useMemo, ReactNode} from 'react'
import {Background, BaseEdge, Controls, Edge, EdgeProps, Handle, MarkerType, Node, NodeProps, Position, ReactFlow, useEdgesState, useNodesState,} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import {useMigrationStore} from '../store/migrationStore'
import type {Migration, Operation, Warning} from '../api/migrationApi'

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

function formatOperationSummary(operation: Operation): ReactNode {
    if (operation.type === 'CREATE_TABLE') {
        if (operation.partitionOf) {
            return `partition(of:${operation.partitionOf})`
        }
        const suffix = operation.isPartitioned ? ' [partitioned]' : ''
        return (
            <div style={{textAlign: 'left'}}>
                <div>create({suffix}</div>
                {operation.columns.map((c) => (
                    <div key={c.name} style={{paddingLeft: 16}}>{c.name}</div>
                ))}
                <div>)</div>
            </div>
        )
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

const X_GAP = 10
const Y_GAP = 5

interface NodeDimensions {
    width: number
    height: number
}

function getOperationTextWidth(operation: Operation): number {
    const CHAR_WIDTH = 5 // ~5px per char at 9px font
    switch (operation.type) {
        case 'CREATE_TABLE':
            if (operation.partitionOf) {
                return `partition(of:${operation.partitionOf})`.length * CHAR_WIDTH
            }
            const colWidths = operation.columns.map(c => c.name.length + 4)
            const maxColWidth = colWidths.length > 0 ? Math.max(...colWidths) : 0
            return Math.max(operation.tableName.length, maxColWidth, 8) * CHAR_WIDTH
        case 'ALTER_COLUMN_TYPE':
            return `type(${operation.columnName}→${operation.newType})`.length * CHAR_WIDTH
        case 'SET_NOT_NULL':
            return `notNull(${operation.columnName})`.length * CHAR_WIDTH
        case 'DROP_NOT_NULL':
            return `nullable(${operation.columnName})`.length * CHAR_WIDTH
        case 'SET_DEFAULT':
            return `default(${operation.columnName}=${operation.defaultValue})`.length * CHAR_WIDTH
        case 'DROP_DEFAULT':
            return `dropDefault(${operation.columnName})`.length * CHAR_WIDTH
        case 'RENAME_TABLE':
            return `rename(→${operation.newTableName})`.length * CHAR_WIDTH
        case 'RENAME_COLUMN':
            return `rename(${operation.columnName}→${operation.newColumnName})`.length * CHAR_WIDTH
        case 'ADD_CONSTRAINT':
            return `constraint(${operation.constraintName}:${operation.constraintType})`.length * CHAR_WIDTH
        case 'DROP_CONSTRAINT':
            return `dropConstraint(${operation.constraintName})`.length * CHAR_WIDTH
        case 'DROP_TABLE':
            return 'drop()'.length * CHAR_WIDTH
        case 'DROP_COLUMN':
            return `dropColumn(${operation.columnName})`.length * CHAR_WIDTH
        case 'ADD_COLUMN':
            return `add(${operation.column.name})`.length * CHAR_WIDTH
        default:
            return 50
    }
}

function calculateNodeSize(operation: Operation, tableName: string, version: string): NodeDimensions {
    const CHAR_WIDTH = 5 // ~5px per char at 9px font
    const LINE_HEIGHT = 11
    const PADDING = 5

    const operationWidth = getOperationTextWidth(operation)
    const tableNameWidth = tableName.length * CHAR_WIDTH
    const versionWidth = version.length * CHAR_WIDTH

    const maxTextWidth = Math.max(operationWidth, tableNameWidth, versionWidth)
    let lineCount = 3 // header + version + operation type

    if (operation.type === 'CREATE_TABLE' && !operation.partitionOf && operation.columns.length > 0) {
        lineCount = 3 + operation.columns.length
    }

    return {
        width: maxTextWidth + PADDING,
        height: lineCount * LINE_HEIGHT + PADDING
    }
}

interface OperationNodeData extends Record<string, unknown> {
    label: ReactNode
    background: string
    hasSource: boolean
    hasTarget: boolean
    warnings: Warning[]
}

function OperationNode({data}: NodeProps<Node<OperationNodeData>>) {
    const hasWarnings = data.warnings.length > 0
    const warningMessage = hasWarnings ? data.warnings.map(w => w.message).join('\n') : undefined

    return (
        <div
            style={{
                background: data.background,
                color: 'white',
                padding: 4,
                borderRadius: 4,
                fontSize: '9px',
                textAlign: 'left',
                border: hasWarnings ? '2px solid #ed8936' : undefined,
                position: 'relative',
            }}
            title={warningMessage}
        >
            {hasWarnings && (
                <div style={{
                    position: 'absolute',
                    top: -6,
                    right: -6,
                    width: 14,
                    height: 14,
                    background: '#ed8936',
                    borderRadius: '50%',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: '10px',
                    fontWeight: 'bold',
                }}>
                    ⚠
                </div>
            )}
            {data.hasTarget && (
                <Handle
                    type="target"
                    position={Position.Left}
                    style={{background: '#555'}}
                />
            )}
            {data.label}
            {data.hasSource && (
                <Handle
                    type="source"
                    position={Position.Right}
                    style={{background: '#555'}}
                />
            )}
        </div>
    )
}

function StepDownEdge({sourceX, sourceY, targetX, targetY, markerEnd, style}: EdgeProps) {
    // Start from right side of source, go down, then right to target
    const path = `M ${sourceX} ${sourceY} L ${sourceX} ${targetY} L ${targetX} ${targetY}`
    return <BaseEdge path={path} markerEnd={markerEnd} style={style}/>
}

function buildWarningMap(warnings: Warning[]): Map<string, Warning[]> {
    const map = new Map<string, Warning[]>()
    for (const warning of warnings) {
        const key = `${warning.operationId.migrationId}-${warning.operationId.tableName}`
        const existing = map.get(key) ?? []
        existing.push(warning)
        map.set(key, existing)
    }
    return map
}

export function Timeline() {
    const {migrations, analysis, loading, error} = useMigrationStore()

    const buildEdgeData = useCallback(() => {
        const {tableOperations, renames, partitions} = buildTableOperationsMap(migrations)
        const edges: Edge[] = []
        const sources = new Set<string>()
        const targets = new Set<string>()

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
                sources.add(sourceId)
                targets.add(targetId)
            }
        })

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
                    style: {strokeDasharray: '5,5'},
                })
                sources.add(sourceId)
                targets.add(targetId)
            }
        })

        partitions.forEach(partition => {
            const parentOps = tableOperations.get(partition.parentTable)
            if (!parentOps || parentOps.length === 0) return
            const parentCreateOp = parentOps.find(op => op.operation.type === 'CREATE_TABLE')
            if (!parentCreateOp) return
            const sourceId = `${parentCreateOp.migration.id}-${partition.parentTable}`
            const targetId = `${partition.migrationId}-${partition.childTable}`
            edges.push({
                id: `partition-${sourceId}->${targetId}`,
                source: sourceId,
                target: targetId,
                type: 'stepDown',
                markerEnd: {type: MarkerType.ArrowClosed},
                style: {strokeDasharray: '3,3', stroke: '#38a169'},
            })
            sources.add(sourceId)
            targets.add(targetId)
        })

        return {edges, sources, targets}
    }, [migrations])

    const warningMap = useMemo(() => buildWarningMap(analysis?.warnings ?? []), [analysis])

    const buildNodes = useCallback((sources: Set<string>, targets: Set<string>): Node[] => {
        const {tableOperations, tableOrder} = buildTableOperationsMap(migrations)

        const uniqueTimestamps = [...new Set(migrations.map(m => m.timestamp))].sort((a, b) => a - b)
        const timestampRank = new Map(uniqueTimestamps.map((ts, idx) => [ts, idx]))

        // Phase 1: Calculate all node sizes and collect node data
        const nodeSizes = new Map<string, NodeDimensions>()
        const nodeData: Array<{
            item: OperationEntry
            tableName: string
            rank: number
            rowIndex: number
            nodeId: string
        }> = []

        tableOrder.forEach((tableName, rowIndex) => {
            const ops = tableOperations.get(tableName) ?? []
            ops.forEach((item) => {
                const nodeTableKey = item.operation.type === 'RENAME_TABLE'
                    ? item.operation.newTableName
                    : item.operation.tableName
                const nodeId = `${item.migration.id}-${nodeTableKey}`
                const rank = timestampRank.get(item.migration.timestamp) ?? 0

                const size = calculateNodeSize(item.operation, tableName, item.migration.version)
                nodeSizes.set(nodeId, size)
                nodeData.push({item, tableName, rank, rowIndex, nodeId})
            })
        })

        // Phase 2: Calculate column widths (max width per rank)
        const columnWidths = new Map<number, number>()
        nodeData.forEach(({rank, nodeId}) => {
            const size = nodeSizes.get(nodeId)!
            const current = columnWidths.get(rank) ?? 0
            columnWidths.set(rank, Math.max(current, size.width))
        })

        // Phase 3: Calculate row heights (max height per row)
        const rowHeights = new Map<number, number>()
        nodeData.forEach(({rowIndex, nodeId}) => {
            const size = nodeSizes.get(nodeId)!
            const current = rowHeights.get(rowIndex) ?? 0
            rowHeights.set(rowIndex, Math.max(current, size.height))
        })

        // Phase 4: Calculate cumulative positions
        const columnX = new Map<number, number>()
        let x = 0
        const maxCol = columnWidths.size > 0 ? Math.max(...columnWidths.keys()) : 0
        for (let col = 0; col <= maxCol; col++) {
            columnX.set(col, x)
            x += (columnWidths.get(col) ?? 0) + X_GAP
        }

        const rowY = new Map<number, number>()
        let y = 0
        const maxRow = rowHeights.size > 0 ? Math.max(...rowHeights.keys()) : 0
        for (let row = 0; row <= maxRow; row++) {
            rowY.set(row, y)
            y += (rowHeights.get(row) ?? 0) + Y_GAP
        }

        // Phase 5: Create nodes with positions and explicit dimensions
        // Center nodes vertically within their row for straight horizontal edges
        return nodeData.map(({item, tableName, rank, rowIndex, nodeId}) => {
            const size = nodeSizes.get(nodeId)!
            const rowHeight = rowHeights.get(rowIndex) ?? size.height
            const verticalOffset = (rowHeight - size.height) / 2
            const nodeWarnings = warningMap.get(nodeId) ?? []

            return {
                id: nodeId,
                type: 'operation',
                position: {
                    x: columnX.get(rank)!,
                    y: rowY.get(rowIndex)! + verticalOffset
                },
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
                    background: getNodeColor(item.operation),
                    hasSource: sources.has(nodeId),
                    hasTarget: targets.has(nodeId),
                    warnings: nodeWarnings,
                },
            }
        })
    }, [migrations, warningMap])

    const [nodes, setNodes] = useNodesState<Node>([])
    const [edges, setEdges] = useEdgesState<Edge>([])
    const edgeTypes = useMemo(() => ({stepDown: StepDownEdge}), [])
    const nodeTypes = useMemo(() => ({operation: OperationNode}), [])

    useEffect(() => {
        const {edges: edgeList, sources, targets} = buildEdgeData()
        setNodes(buildNodes(sources, targets))
        setEdges(edgeList)
    }, [migrations, buildNodes, buildEdgeData, setNodes, setEdges])

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
                nodeTypes={nodeTypes}
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
