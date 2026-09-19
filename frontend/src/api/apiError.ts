export class ApiError extends Error {
  constructor(message: string, readonly status: number, readonly requestId: string | null) {
    super(message)
    this.name = 'ApiError'
  }
}
