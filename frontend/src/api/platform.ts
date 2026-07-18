import request from './request'
import type { ApiResponse, Platform } from '@/types/platform'

/** 获取所有平台配置（含菜单/页面/字段定义） */
export function getPlatforms() {
  return request.get<any, ApiResponse<Platform[]>>('/platforms')
}

/** 获取单个平台详情 */
export function getPlatform(id: string) {
  return request.get<any, ApiResponse<Platform>>(`/platforms/${id}`)
}

/** 新增平台 */
export function createPlatform(data: Partial<Platform>) {
  return request.post<any, ApiResponse<Platform>>('/platforms', data)
}

/** 更新平台 */
export function updatePlatform(id: string, data: Partial<Platform>) {
  return request.put<any, ApiResponse<Platform>>(`/platforms/${id}`, data)
}

/** 删除平台 */
export function deletePlatform(id: string) {
  return request.delete<any, ApiResponse<null>>(`/platforms/${id}`)
}
