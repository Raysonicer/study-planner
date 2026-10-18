import { ElMessage } from 'element-plus'
import { LocalError } from '@/local/errors'
import { exportAll, importAll, type ExportPayload } from '@/local/export'

// 纯前端版 api 层：本地服务调用统一包装为 axios 风格的响应/错误形状，
// 使页面、store、组件的 `.data` 与 catch 逻辑保持零改动。

export interface ApiResponse<T> {
  data: T
}

export interface HttpError extends Error {
  response?: {
    status: number
    data: { code: string; message: string }
  }
  config?: { skipErrorToast?: boolean }
}

export interface LocalCallOpts {
  /** 预期内的错误（如探测资源是否存在）不弹错误提示 */
  skipErrorToast?: boolean
}

function toHttpError(e: unknown, opts?: LocalCallOpts): HttpError {
  if (e instanceof LocalError) {
    const status = e.code === 'not_found' ? 404 : e.code === 'conflict' ? 409 : 400
    const err = new Error(e.message) as HttpError
    err.response = { status, data: { code: e.code, message: e.message } }
    err.config = { skipErrorToast: opts?.skipErrorToast }
    return err
  }
  return e as HttpError
}

/** 包装本地服务调用：成功返回 { data }，失败抛 axios 风格错误并弹提示。 */
export async function local<T>(fn: () => Promise<T>, opts?: LocalCallOpts): Promise<ApiResponse<T>> {
  try {
    return { data: await fn() }
  } catch (e) {
    const err = toHttpError(e, opts)
    if (!opts?.skipErrorToast) {
      const msg = err.response?.data?.message || err.message || '操作失败'
      ElMessage.error(msg)
    }
    throw err
  }
}

/**
 * 兼容 shim：设置页的导出/导入仍走 http 风格调用。
 * 纯前端版中导出即下载 JSON，导入即整体替换本地数据。
 */
export const http = {
  async get(path: string, _opts?: unknown): Promise<ApiResponse<Blob>> {
    if (path === '/export') {
      const payload = await exportAll()
      const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' })
      return { data: blob }
    }
    throw new Error(`未知请求 GET ${path}`)
  },
  async post(path: string, body?: unknown): Promise<ApiResponse<unknown>> {
    if (path === '/export/import') {
      await importAll(body as ExportPayload)
      return { data: { imported: true } }
    }
    throw new Error(`未知请求 POST ${path}`)
  },
}

