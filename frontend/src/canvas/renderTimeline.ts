import {CreateTable, RenameTable} from '../api/migrationApi'
import {LayoutInfo, OperationLayoutInfo, TRANSITION_TAG_SHIFT} from './layoutInfo.ts'
import {drawOperation} from './drawOperation.ts'
import {drawTransitionRibbon, RibbonInfo} from "./drawTransitionRibbon.ts";
import {getColorFromString} from "./getColorFromString.ts";
import {drawMigrationRect} from "./drawMigrationRect.ts";

// On-screen pixel thresholds below which texts and flags are hidden.
const FONT_SIZE = 13
const MIN_TEXT_PX = 7
const FLAG_HEIGHT = FONT_SIZE + 6 // world height of a flag, see drawFlag
const MIN_FLAG_PX = 5

// Paints the precomputed layout onto a viewport-sized canvas, uniformly scaled
// so the whole timeline fits the viewport and centered in the leftover space.
// Texts and flags are hidden once the scale makes them unreadable.
// Accounts for device pixel ratio.
export function renderTimeline(
    canvas: HTMLCanvasElement,
    layout: LayoutInfo,
    viewport: { width: number; height: number },
    scale: number,
): void {
    const dpr = window.devicePixelRatio || 1
    const backingWidth = Math.round(viewport.width * dpr)
    const backingHeight = Math.round(viewport.height * dpr)

    // Only resize when needed; reassigning canvas.width/height clears and
    // reallocates the backing store.
    if (canvas.width !== backingWidth || canvas.height !== backingHeight) {
        canvas.width = backingWidth
        canvas.height = backingHeight
        canvas.style.width = `${viewport.width}px`
        canvas.style.height = `${viewport.height}px`
    }

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.clearRect(0, 0, viewport.width, viewport.height)
    ctx.translate((viewport.width - layout.width * scale) / 2, (viewport.height - layout.height * scale) / 2)
    ctx.scale(scale, scale)

    const showText = FONT_SIZE * scale >= MIN_TEXT_PX
    const showFlags = FLAG_HEIGHT * scale >= MIN_FLAG_PX

    // Rename cells keyed by the old table name; partition children keyed by parent table name.
    let renamedTables = new Map<string, { info: OperationLayoutInfo, rename: RenameTable }>()
    let partitionedTables = new Map<string, { info: OperationLayoutInfo, tableName: string }[]>()
    let transitionRibbons: RibbonInfo[] = []
    const migrationsToDraw: {
        op: OperationLayoutInfo;
        prevMig: OperationLayoutInfo | null;
        nextMig: OperationLayoutInfo | null;
        renderGradient: boolean
    }[] = []

    // Painted last table first, and within each table last migration first,
    // so rendering starts from the last migration of the last table.
    const entries = [...layout.tableMigrations]
    for (let t = entries.length - 1; t >= 0; t--) {
        const [table, tableOperations] = entries[t]

        // Row label (table name) in the left gutter
        if (showText) {
            ctx.fillStyle = getColorFromString(table.tableName)
            ctx.font = '13px sans-serif'
            ctx.textAlign = 'left'
            ctx.textBaseline = 'middle'
            ctx.fillText(table.tableName, 8, table.y + table.h / 2)
        }

        for (let m = tableOperations.length - 1; m >= 0; m--) {
            const nextOp = tableOperations[m + 1] ?? null
            const prevOp = tableOperations[m - 1] ?? null
            const currentOp = tableOperations[m];
            const ops = currentOp.operations
            const renameOp = ops.find((o): o is RenameTable => o.type === 'RENAME_TABLE')
            const partitionChildOp = ops.find((o): o is CreateTable =>
                o.type === 'CREATE_TABLE' && o.partitionOf != null)
            let renderGradient = true
            if (renameOp) {
                renamedTables.set(renameOp.tableName, {info: currentOp, rename: renameOp})
            } else if (partitionChildOp) {
                const children = partitionedTables.get(partitionChildOp.partitionOf!) ?? []
                children.push({info: currentOp, tableName: partitionChildOp.tableName})
                partitionedTables.set(partitionChildOp.partitionOf!, children)
            } else {
                const tableName = ops[0].tableName
                if (renamedTables.has(tableName)) {
                    renderGradient = false
                    const renamed = renamedTables.get(tableName)!!
                    renamedTables.delete(tableName)

                    const r: RibbonInfo = {
                        leftX: renamed.info.x,
                        leftY: renamed.info.y + TRANSITION_TAG_SHIFT,
                        rightX: currentOp.x,
                        rightY: currentOp.y + TRANSITION_TAG_SHIFT,
                        height: currentOp.h - TRANSITION_TAG_SHIFT,
                        leftColor: getColorFromString(renamed.rename.newTableName),
                        rightColor: getColorFromString(tableName),
                        init: prevOp === null
                    }
                    transitionRibbons.push(r)
                }
            }

            // Partitioned parent: connect it to each stored child partition.
            const partitionParentOp = ops.find((o): o is CreateTable =>
                o.type === 'CREATE_TABLE' && o.isPartitioned)
            if (partitionParentOp && partitionedTables.has(partitionParentOp.tableName)) {
                renderGradient = false
                const children = partitionedTables.get(partitionParentOp.tableName)!!
                partitionedTables.delete(partitionParentOp.tableName)

                children.forEach(child => {
                    transitionRibbons.push({
                        leftX: child.info.x,
                        leftY: child.info.y + TRANSITION_TAG_SHIFT,
                        rightX: currentOp.x,
                        rightY: currentOp.y + TRANSITION_TAG_SHIFT,
                        height: currentOp.h - TRANSITION_TAG_SHIFT,
                        leftColor: getColorFromString(child.tableName),
                        rightColor: getColorFromString(partitionParentOp.tableName),
                        init: prevOp === null
                    })
                })
            }
            migrationsToDraw.push({op: currentOp, prevMig: prevOp, nextMig: nextOp, renderGradient})
        }
    }

    // Migration outlines behind everything, then ribbons, then operations on
    // top, so boxes are never overlapped by ribbons.
    layout.migrationBounds.forEach(b => drawMigrationRect(ctx, b, scale))
    transitionRibbons.forEach(r => drawTransitionRibbon(ctx, r))
    migrationsToDraw.forEach(m => drawOperation(ctx, m.op, m.prevMig, m.nextMig, m.renderGradient, showFlags, showText))
}
