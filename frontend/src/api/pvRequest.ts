/**
 * 光伏平台专用请求实例
 * ----------------------------------------------------------------
 * 原光伏项目的自定义组件（DashboardView、DeviceDetailView 等）调用的是
 * 原项目专用 API 路径（/stats/overview、/devices/tree、/events 等），
 * 这些接口在新架构后端中尚未实现。
 *
 * 为了让原光伏组件在新架构中直接可用，此请求实例使用 mock 适配器，
 * 通过 pvMock.ts 返回演示数据。后续接入真实后端时，只需移除 adapter
 * 并将路径加上 /pv/ 前缀即可。
 */
import axios from 'axios'
import { matchPvMock } from './pvMock'

const pvRequest = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// 请求拦截器：附加 token
pvRequest.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 自定义 adapter：所有光伏专用 API 走 mock 数据
pvRequest.defaults.adapter = async (config: any) => {
  const method = (config.method || 'get').toLowerCase()
  const url = (config.url || '').split('?')[0]
  const mockData = matchPvMock(method, url, config.params)

  return {
    data: { code: 0, data: mockData, message: 'ok' },
    status: 200,
    statusText: 'OK',
    headers: {},
    config,
    request: {}
  } as any
}

// 响应拦截器：解包 response.data，返回 ApiResponse 结构
pvRequest.interceptors.response.use((response) => {
  return response.data
})

export default pvRequest
