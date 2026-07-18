import request from './request'
import type { ApiResponse, UserInfo } from '@/types/platform'

/** 用户登录 */
export function login(username: string, password: string) {
  return request.post<any, ApiResponse<{ token: string; user: UserInfo }>>(
    '/auth/login',
    { username, password }
  )
}

/** 获取当前用户信息 */
export function getProfile() {
  return request.get<any, ApiResponse<UserInfo>>('/auth/profile')
}

/** 退出登录 */
export function logout() {
  return request.post<any, ApiResponse<null>>('/auth/logout')
}
