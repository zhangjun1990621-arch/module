import request from './request'
import type { ApiResponse, PageResult } from '@/types/platform'

/**
 * 通用资源 API
 * ------------------------------------------------
 * 所有平台的数据接口统一遵循 RESTful 约定：
 *   GET    /api/:platform/:resource       → 列表查询（支持分页/筛选参数）
 *   POST   /api/:platform/:resource       → 新建资源
 *   PUT    /api/:platform/:resource/:id   → 更新资源
 *   DELETE /api/:platform/:resource/:id   → 删除资源
 *
 * 前端不需要知道具体有哪些平台和资源，一切由后端配置驱动。
 */

/** 查询资源列表 */
export function getResources(
  platform: string,
  resource: string,
  params?: Record<string, any>
) {
  return request.get<any, ApiResponse<PageResult | any[]>>(
    `/${platform}/${resource}`,
    { params }
  )
}

/** 新建资源 */
export function createResource(
  platform: string,
  resource: string,
  data: Record<string, any>
) {
  return request.post<any, ApiResponse<any>>(`/${platform}/${resource}`, data)
}

/** 更新资源 */
export function updateResource(
  platform: string,
  resource: string,
  id: string | number,
  data: Record<string, any>
) {
  return request.put<any, ApiResponse<any>>(`/${platform}/${resource}/${id}`, data)
}

/** 删除资源 */
export function deleteResource(
  platform: string,
  resource: string,
  id: string | number
) {
  return request.delete<any, ApiResponse<null>>(`/${platform}/${resource}/${id}`)
}
