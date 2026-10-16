import axios from 'axios'
import { ElMessage } from 'element-plus'

declare module 'axios' {
  export interface AxiosRequestConfig {
    /** 预期内的错误（如探测资源是否存在）不弹错误提示 */
    skipErrorToast?: boolean
  }
}

export const http = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

http.interceptors.response.use(
  (res) => res,
  (err) => {
    if (!err.config?.skipErrorToast) {
      const msg = err.response?.data?.message || err.message || '请求失败'
      ElMessage.error(msg)
    }
    return Promise.reject(err)
  },
)