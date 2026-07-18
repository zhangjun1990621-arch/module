import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getPlatforms as apiGetPlatforms } from '@/api/platform'
import { useAuthStore } from './auth'
import type { Platform } from '@/types/platform'

const STORAGE_KEY = 'iot_current_platform'

export const usePlatformStore = defineStore('platform', () => {
  const authStore = useAuthStore()

  /** 所有平台列表（从后端动态加载） */
  const platforms = ref<Platform[]>([])

  /** 当前选中的平台 */
  const currentPlatform = ref<Platform | null>(null)

  /** 是否已加载平台数据 */
  const loaded = ref(false)

  /** 当前用户可见的平台列表（基于权限过滤） */
  const visiblePlatforms = computed(() => {
    return platforms.value.filter((p) => {
      if (p.status === 'disabled') return false
      return authStore.canAccessPlatform(p.id)
    })
  })

  /** 从后端加载所有平台配置 */
  async function loadPlatforms() {
    const res = await apiGetPlatforms()
    platforms.value = (res.data || []).sort((a, b) => a.sortOrder - b.sortOrder)
    loaded.value = true

    // 恢复上次选中的平台，或默认选第一个可见平台
    const savedId = localStorage.getItem(STORAGE_KEY)
    const saved = savedId
      ? platforms.value.find((p) => p.id === savedId)
      : null
    if (saved && authStore.canAccessPlatform(saved.id)) {
      currentPlatform.value = saved
    } else if (visiblePlatforms.value.length > 0) {
      currentPlatform.value = visiblePlatforms.value[0]
      localStorage.setItem(STORAGE_KEY, currentPlatform.value.id)
    }
    return platforms.value
  }

  /** 切换当前平台 */
  function selectPlatform(id: string) {
    const target = platforms.value.find((p) => p.id === id)
    if (!target) return
    if (!authStore.canAccessPlatform(id)) return
    currentPlatform.value = target
    localStorage.setItem(STORAGE_KEY, id)
  }

  /** 根据 ID 获取平台 */
  function getPlatformById(id: string): Platform | undefined {
    return platforms.value.find((p) => p.id === id)
  }

  /** 重置 store（退出登录时调用） */
  function reset() {
    platforms.value = []
    currentPlatform.value = null
    loaded.value = false
  }

  return {
    platforms,
    currentPlatform,
    loaded,
    visiblePlatforms,
    loadPlatforms,
    selectPlatform,
    getPlatformById,
    reset
  }
})
