// 本地存储层的错误类型，与 Go 版 service 层错误一一对应
export type ErrorCode = 'not_found' | 'invalid_input' | 'conflict'

export class LocalError extends Error {
  code: ErrorCode
  constructor(code: ErrorCode, message: string) {
    super(message)
    this.name = 'LocalError'
    this.code = code
  }
}

export const errNotFound = (msg = '资源不存在') => new LocalError('not_found', msg)
export const errInvalid = (msg: string) => new LocalError('invalid_input', msg)
export const errConflict = (msg: string) => new LocalError('conflict', msg)
