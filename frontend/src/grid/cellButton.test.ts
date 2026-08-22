import { describe, expect, it } from 'vitest'
import type { CellContext } from 'active-grid'
import { buttonRect, buttonsDecorHeight } from './drawers/cellButton'

function decorCell(slots: number): CellContext {
  return {
    rect: { x: 0, y: 0, width: 400, height: buttonsDecorHeight(slots) },
  } as CellContext
}

describe('footer buttons', () => {
  it('stacks slots top-down without overlapping', () => {
    const cell = decorCell(2)

    const focus = buttonRect(cell, 0)
    const runtime = buttonRect(cell, 1)

    expect(runtime.y).toBeGreaterThanOrEqual(focus.y + focus.height)
    expect(runtime.x).toBe(focus.x)
  })

  it('keeps every slot inside the decor it sizes', () => {
    const slots = 2
    const cell = decorCell(slots)

    for (let slot = 0; slot < slots; slot++) {
      const rect = buttonRect(cell, slot)
      expect(rect.y).toBeGreaterThanOrEqual(cell.rect.y)
      expect(rect.y + rect.height).toBeLessThanOrEqual(cell.rect.y + cell.rect.height)
    }
  })

  it('grows the decor with each added button', () => {
    expect(buttonsDecorHeight(2)).toBeGreaterThan(buttonsDecorHeight(1))
  })
})
