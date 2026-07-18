import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import type { Platform, NavItem } from '@/types/platform'
import { useAuthStore } from '@/stores/auth'
import { usePlatformStore } from '@/stores/platform'

/**
 * 静态路由定义
 * ------------------------------------------------
 * 只有登录页、首页概览、平台管理是静态注册的。
 * 所有平台的业务页面均通过 setupDynamicRoutes() 动态注册。
 */
const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/LoginView.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/components/MainLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: '', redirect: '/dashboard' },
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/DashboardView.vue')
      },
      {
        path: 'platform-management',
        name: 'PlatformManagement',
        component: () => import('@/views/PlatformManagement.vue'),
        meta: { requiresSuperAdmin: true }
      },
      {
        path: 'permission-management',
        name: 'PermissionManagement',
        component: () => import('@/views/PermissionManagement.vue'),
        meta: { requiresSuperAdmin: true }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

/** 动态路由是否已注册 */
let dynamicRoutesAdded = false

/**
 * 动态路由注册
 * ------------------------------------------------
 * 遍历所有平台的 navItems，对每个菜单项调用 router.addRoute() 注册路由。
 * 路由 component 统一指向 DynamicPage 通用页面组件，
 * 通过 props 传入 platformId 和 pagePath，由 DynamicPage 根据平台配置渲染对应内容。
 *
 * 新增平台零前端代码改动 —— 只需后端在 /api/platforms 返回新平台配置即可。
 */
export function setupDynamicRoutes(platforms: Platform[]) {
  if (dynamicRoutesAdded) return

  for (const platform of platforms) {
    const navItems = platform.config?.navItems || []
    if (navItems.length === 0) continue

    // 找到第一个有 path 的菜单项作为平台默认跳转目标
    const firstPath = findFirstPath(navItems)

    // 注册平台根路径重定向
    if (firstPath) {
      router.addRoute('Layout', {
        path: platform.id,
        redirect: firstPath
      })
    }

    // 注册所有菜单项的路由（含子菜单）
    registerNavRoutes(navItems)
  }

  // 兜底：未匹配的路由重定向到首页
  router.addRoute({
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    redirect: '/dashboard'
  })

  dynamicRoutesAdded = true
}

/** 递归查找第一个有 path 的菜单项（含子菜单） */
function findFirstPath(navItems: NavItem[]): string | undefined {
  for (const item of navItems) {
    if (item.path) return item.path
    if (item.children?.length) {
      const childPath = findFirstPath(item.children)
      if (childPath) return childPath
    }
  }
  return undefined
}

/** 递归注册菜单项路由（支持二级菜单） */
function registerNavRoutes(navItems: NavItem[]) {
  for (const item of navItems) {
    if (item.path) {
      // 有 path 的菜单项注册路由
      const relativePath = item.path.replace(/^\//, '')
      const segments = relativePath.split('/')
      const platformId = segments[0]
      const pagePath = segments.slice(1).join('/')

      if (pagePath) {
        router.addRoute('Layout', {
          path: relativePath,
          name: `dynamic_${platformId}_${pagePath.replace(/\//g, '_')}`,
          component: () => import('@/components/DynamicPage.vue'),
          props: { platformId, pagePath },
          meta: { platformId, pagePath }
        })
      }
    }
    // 递归注册子菜单
    if (item.children?.length) {
      registerNavRoutes(item.children)
    }
  }
}

/** 重置动态路由（退出登录时调用） */
export function resetDynamicRoutes() {
  dynamicRoutesAdded = false
}

/**
 * 全局路由守卫
 * ------------------------------------------------
 * 1. 未登录 → 跳转 /login
 * 2. 已登录但平台数据未加载 → 先 loadPlatforms() 再 setupDynamicRoutes()
 * 3. 平台管理页仅 super_admin 可访问
 */
router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()
  const platformStore = usePlatformStore()

  // 登录页不需要鉴权
  if (to.path === '/login') {
    next()
    return
  }

  // 未登录跳转登录页
  if (!authStore.token) {
    next('/login')
    return
  }

  // 恢复用户信息
  if (!authStore.user) {
    authStore.loadUser()
  }

  // 平台管理页权限校验
  if (to.meta.requiresSuperAdmin && !authStore.isSuperAdmin) {
    next('/dashboard')
    return
  }

  // 已登录但平台数据未加载 → 加载平台 + 注册动态路由
  if (!platformStore.loaded) {
    try {
      await platformStore.loadPlatforms()
      setupDynamicRoutes(platformStore.platforms)
      // 动态路由注册后需重新匹配当前导航目标
      next({ ...to, replace: true })
      return
    } catch {
      authStore.logout()
      platformStore.reset()
      resetDynamicRoutes()
      next('/login')
      return
    }
  }

  next()
})

export default router
