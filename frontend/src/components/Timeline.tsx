import {useCallback, useEffect} from 'react'
import {Background, Controls, Edge, MarkerType, Node, Position, ReactFlow, useEdgesState, useNodesState,} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import {useMigrationStore} from '../store/migrationStore'
import type {Migration, Operation} from '../api/migrationApi'

interface OperationEntry {
    migration: Migration
    operation: Operation
}

function formatOperationSummary(operation: Operation): string {
    if (operation.type === 'CREATE_TABLE') {
        const columnNames = operation.columns.map(c => c.name).join(', ')
        return `create(${columnNames})`
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
    if (operation.droppedColumns.length > 0) {
        return '#c05621' // orange
    }
    return '#276749' // green
}

function buildTableOperationsMap(migrations: Migration[]): Map<string, OperationEntry[]> {
    const tableOperations = new Map<string, OperationEntry[]>()

    migrations.forEach(migration => {
        migration.operations.forEach(operation => {
            const tableName = operation.tableName
            if (!tableOperations.has(tableName)) {
                tableOperations.set(tableName, [])
            }
            tableOperations.get(tableName)!.push({migration, operation})
        })
    })

    return tableOperations
}

const NODE_WIDTH = 180
const NODE_HEIGHT = 80
const X_GAP = 40
const Y_GAP = 30

export function Timeline() {
    const {migrations, loading, error} = useMigrationStore()

    const buildNodes = useCallback((): Node[] => {
        const tableOperations = buildTableOperationsMap(migrations)
        const nodes: Node[] = []

        const uniqueTimestamps = [...new Set(migrations.map(m => m.timestamp))].sort((a, b) => a - b)
        const timestampRank = new Map(uniqueTimestamps.map((ts, idx) => [ts, idx]))

        let rowIndex = 0
        tableOperations.forEach((ops, tableName) => {
            ops.forEach((item) => {
                const nodeId = `${item.migration.id}-${item.operation.tableName}`
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
            rowIndex++
        })

        return nodes
    }, [migrations])

    const buildEdges = useCallback((): Edge[] => {
        const tableOperations = buildTableOperationsMap(migrations)
        const edges: Edge[] = []

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
