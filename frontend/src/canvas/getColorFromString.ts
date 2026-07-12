// Maps a string to a stable, visually distinct color.
// Same input always yields the same color; used to give each table its own hue.
export function getColorFromString(value: string): string {
    let hash = 0
    for (let i = 0; i < value.length; i++) {
        hash = value.charCodeAt(i) + ((hash << 5) - hash)
        hash |= 0
    }

    const hue = Math.abs(hash) % 360
    return `hsl(${hue}, 65%, 55%)`
}
