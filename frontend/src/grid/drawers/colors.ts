/** Ink used by every text-drawing layer. */
export const FG = '#e0fbfc';

export function lerpColor(a: string, b: string, t: number): string {
    const pa = parseInt(a.slice(1), 16);
    const pb = parseInt(b.slice(1), 16);
    const ch = (shift: number) => {
        const from = (pa >> shift) & 255;
        const to = (pb >> shift) & 255;
        return Math.round(from + (to - from) * t);
    };
    return `rgb(${ch(16)}, ${ch(8)}, ${ch(0)})`;
}
