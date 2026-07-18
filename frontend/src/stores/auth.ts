import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as apiLogin, getProfile as apiGetProfile, logout as apiLogout } from '@/api/auth'
import type { UserInfo } from '@/types/platform'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<UserInfo | null>(null)

  /** 当前用户可访问的平台列表 */
  const userPlatforms = computed<string[]>(() => {
    if (!user.value?.platforms) return []
    return user.value.platforms.split(',').filter(Boolean)
  })

  /** 是否为超级管理员 */
  const isSuperAdmin = computed(() => user.value?.role === 'super_admin')

  /** 是否为管理员（含超管） */
  const isAdmin = computed(() => ['super_admin', 'admin'].includes(user.value?.role || ''))

  /** 登录 */
  async function login(username: string, password: string) {
    const res = await apiLogin(username, password)
    token.value = res.data.token
    user.value = res.data.user
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('user', JSON.stringify(res.data.user))
    return res
  }

  /** 退出登录 */
  function logout() {
    // 尝试通知后端注销 token（不阻塞前端跳转）
    apiLogout().catch(() => {})
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    localStorage.removeItem('iot_current_platform')
  }

  /** 从 localStorage 恢复用户信息 */
  function loadUser() {
    const saved = localStorage.getItem('user')
    if (saved) {
      try {
        user.value = JSON.parse(saved)
      } catch {
        user.value = null
      }
    }
  }

  /** 判断当前用户是否有权访问指定平台 */
  function canAccessPlatform(platformId: string): boolean {
    if (isSuperAdmin.value) return true
    return userPlatforms.value.includes(platformId)
  }

  return {
    token,
    user,
    userPlatforms,
    isSuperAdmin,
    isAdmin,
    login,
    logout,
    loadUser,
    canAccessPlatform
  }
})
