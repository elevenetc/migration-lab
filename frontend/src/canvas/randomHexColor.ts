export function randomHexColor(): string {
    const color = Math.floor(Math.random() * 0x1000000)
    return `#${color.toString(16).padStart(6, '0')}`
}
