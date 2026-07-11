import {Migration} from '../api/migrationApi'
import {computeLayout, OperationLayoutInfo, TABLE_ROW_HEIGHT, TRANSITION_TAG_SHIFT} from './layoutInfo.ts'
import {drawMigration} from './drawMigration'
import {drawTransitionRibbon, RibbonInfo} from "./drawTransitionRibbon.ts";
import {getOperationColor} from "./getOperationColor.ts";

// Computes the layout for the given migrations and paints it onto the canvas,
// sizing the canvas to the content and accounting for device pixel ratio.
export function renderTimeline(canvas: HTMLCanvasElement, migrations: Migration[]): void {
    const layout = computeLayout(migrations)
    const dpr = window.devicePixelRatio || 1

    canvas.width = layout.width * dpr
    canvas.height = layout.height * dpr
    canvas.style.width = `${layout.width}px`
    canvas.style.height = `${layout.height}px`

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.clearRect(0, 0, layout.width, layout.height)

    let renamedTables = new Map<string, OperationLayoutInfo>()
    let partitionedTables = new Map<string, OperationLayoutInfo[]>()
    let transitionRibbons: RibbonInfo[] = []
    const migrationsToDraw: {
        op: OperationLayoutInfo;
        nextMig: OperationLayoutInfo | null;
        renderGradient: boolean
    }[] = []

    // Painted last table first, and within each table last migration first,
    // so rendering starts from the last migration of the last table.
    const entries = [...layout.tableMigrations]
    for (let t = entries.length - 1; t >= 0; t--) {
        const [table, tableMigrations] = entries[t]

        // Row label (table name) in the left gutter
        ctx.fillStyle = '#e2e8f0'
        ctx.font = '13px sans-serif'
        ctx.textAlign = 'left'
        ctx.textBaseline = 'middle'
        ctx.fillText(table.tableName, 8, table.y + table.h / 2)

        for (let m = tableMigrations.length - 1; m >= 0; m--) {
            const nextMig = tableMigrations[m + 1] ?? null
            const currentOp = tableMigrations[m];
            const op = currentOp.operation
            let renderGradient = true
            if (op.type === 'RENAME_TABLE') {
                renamedTables.set(op.tableName, currentOp)
            } else if (op.type === 'CREATE_TABLE' && op.partitionOf) {
                const children = partitionedTables.get(op.partitionOf) ?? []
                children.push(currentOp)
                partitionedTables.set(op.partitionOf, children)
            } else {
                if (renamedTables.has(op.tableName)) {
                    renderGradient = false
                    const renamedOpLayoutInfo = renamedTables.get(op.tableName)!!
                    renamedTables.delete(op.tableName)

                    const r: RibbonInfo = {
                        leftX: renamedOpLayoutInfo.x,
                        leftY: renamedOpLayoutInfo.y + TRANSITION_TAG_SHIFT,
                        rightX: currentOp.x,
                        rightY: currentOp.y + TRANSITION_TAG_SHIFT,
                        height: TABLE_ROW_HEIGHT - TRANSITION_TAG_SHIFT,
                        leftColor: getOperationColor(renamedOpLayoutInfo.operation),
                        rightColor: getOperationColor(currentOp.operation),
                    }
                    transitionRibbons.push(r)
                }
            }

            // Partitioned parent: connect it to each stored child partition.
            if (op.type === 'CREATE_TABLE' && op.isPartitioned && partitionedTables.has(op.tableName)) {
                renderGradient = false
                const children = partitionedTables.get(op.tableName)!!
                partitionedTables.delete(op.tableName)

                children.forEach(child => {
                    transitionRibbons.push({
                        leftX: child.x,
                        leftY: child.y + TRANSITION_TAG_SHIFT,
                        rightX: currentOp.x,
                        rightY: currentOp.y + TRANSITION_TAG_SHIFT,
                        height: TABLE_ROW_HEIGHT - TRANSITION_TAG_SHIFT,
                        leftColor: getOperationColor(child.operation),
                        rightColor: getOperationColor(currentOp.operation),
                    })
                })
            }
            migrationsToDraw.push({op: currentOp, nextMig, renderGradient})
        }
    }

    // Ribbons first, migrations on top, so boxes are never overlapped by ribbons.
    transitionRibbons.forEach(r => drawTransitionRibbon(ctx, r))
    migrationsToDraw.forEach(m => drawMigration(ctx, m.op, m.nextMig, m.renderGradient))
}
