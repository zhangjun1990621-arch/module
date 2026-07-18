import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'
import type { ApiResponse } from '@/types/platform'

// 扩展 axios 配置，支持 skipErrorMessage 选项（用于 dashboard 等静默请求）
declare module 'axios' {
  interface AxiosRequestConfig {
    skipErrorMessage?: boolean
  }
}

const request = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// 请求拦截器：自动附加 Bearer token
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器：统一处理业务错误码和 401 鉴权失效
// 拦截器解包 response.data，后续调用方直接获取 ApiResponse 结构
request.interceptors.response.use(
  (response) => {
    const data = response.data as ApiResponse
    if (data.code !== 0) {
      ElMessage.error(data.message || '请求失败')
      return Promise.reject(data)
    }
    return data as any
  },
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      router.push('/login')
      ElMessage.error('登录已过期，请重新登录')
    } else if (!error.config?.skipErrorMessage) {
      ElMessage.error(error.message || '网络错误')
    }
    return Promise.reject(error)
  }
)

export default request
