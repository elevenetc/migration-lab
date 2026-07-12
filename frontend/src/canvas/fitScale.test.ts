import { describe, expect, it } from 'vitest'
import { computeFitScale } from './fitScale'

describe('computeFitScale', () => {
    it('returns 1 when layout fits the viewport', () => {
        expect(computeFitScale({ width: 1000, height: 800 }, { width: 500, height: 400 })).toBe(1)
    })

    it('scales down by width when layout is too wide', () => {
        expect(computeFitScale({ width: 500, height: 800 }, { width: 1000, height: 400 })).toBe(0.5)
    })

    it('scales down by height when layout is too tall', () => {
        expect(computeFitScale({ width: 1000, height: 200 }, { width: 500, height: 800 })).toBe(0.25)
    })

    it('uses the smaller ratio when layout exceeds both dimensions', () => {
        expect(computeFitScale({ width: 500, height: 200 }, { width: 1000, height: 800 })).toBe(0.25)
    })
})
