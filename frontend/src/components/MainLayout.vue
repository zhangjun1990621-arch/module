<template>
  <div class="main-layout">
    <!-- ===== 左侧导航栏 ===== -->
    <aside class="sidebar">
      <!-- 平台切换 -->
      <div class="platform-switcher">
        <el-select
          v-model="currentPlatformId"
          placeholder="选择平台"
          @change="onPlatformChange"
          class="platform-select"
        >
          <template #prefix>
            <span class="switcher-icon">{{ currentPlatform?.icon }}</span>
          </template>
          <el-option
            v-for="p in platformStore.visiblePlatforms"
            :key="p.id"
            :label="p.name"
            :value="p.id"
          >
            <span style="margin-right: 8px">{{ p.icon }}</span>
            {{ p.name }}
          </el-option>
        </el-select>
      </div>

      <!-- 导航菜单 -->
      <el-menu
        :default-active="activeMenu"
        :collapse="false"
        class="sidebar-menu"
        @select="onMenuSelect"
      >
        <!-- 静态菜单：首页概览 -->
        <el-menu-item index="/dashboard">
          <el-icon><Odometer /></el-icon>
          <span>首页概览</span>
        </el-menu-item>

        <!-- 动态菜单：当前平台的 navItems（支持二级菜单） -->
        <el-menu-item-group
          v-if="currentPlatform && currentPlatform.config?.navItems?.length"
          :title="currentPlatform.name"
        >
          <template v-for="item in currentPlatform.config.navItems" :key="item.path || item.label">
            <!-- 二级菜单：有 children 时渲染为可折叠的子菜单 -->
            <el-sub-menu v-if="item.children && item.children.length" :index="item.label">
              <template #title>
                <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
                <span>{{ item.label }}</span>
              </template>
              <el-menu-item
                v-for="child in item.children"
                :key="child.path"
                :index="child.path"
              >
                <el-icon v-if="child.icon"><component :is="child.icon" /></el-icon>
                <span>{{ child.label }}</span>
              </el-menu-item>
            </el-sub-menu>

            <!-- 一级菜单：无 children 时直接渲染为菜单项 -->
            <el-menu-item v-else :index="item.path">
              <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
              <span>{{ item.label }}</span>
            </el-menu-item>
          </template>
        </el-menu-item-group>

        <!-- 静态菜单：平台管理（仅超管可见） -->
        <el-menu-item v-if="authStore.isSuperAdmin" index="/platform-management">
          <el-icon><Setting /></el-icon>
          <span>平台管理</span>
        </el-menu-item>

        <!-- 静态菜单：权限管理（仅超管可见） -->
        <el-menu-item v-if="authStore.isSuperAdmin" index="/permission-management">
          <el-icon><Key /></el-icon>
          <span>权限管理</span>
        </el-menu-item>
      </el-menu>
    </aside>

    <!-- ===== 右侧主区域 ===== -->
    <div class="main-area">
      <!-- 顶部栏 -->
      <header class="app-header">
        <div class="header-left">
          <span class="current-platform-name">
            {{ currentPlatform?.icon }} {{ currentPlatform?.name || '综合能源云控平台' }}
          </span>
        </div>
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <div class="user-info">
              <div class="avatar">
                {{ authStore.user?.username?.[0]?.toUpperCase() || 'U' }}
              </div>
              <span class="username">{{ authStore.user?.username || '未知' }}</span>
              <el-tag size="small" type="info" effect="dark">{{ roleLabel }}</el-tag>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <!-- 内容区域 -->
      <main class="page-content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { usePlatformStore } from '@/stores/platform'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const platformStore = usePlatformStore()

authStore.loadUser()

const currentPlatform = computed(() => platformStore.currentPlatform)
const currentPlatformId = ref(currentPlatform.value?.id || '')

/** 当前激活的菜单项（支持二级菜单路径匹配） */
const activeMenu = computed(() => {
  // 精确匹配当前路径
  if (route.path === '/dashboard' || route.path === '/platform-management' || route.path === '/permission-management') {
    return route.path
  }
  // 动态路由：递归匹配当前平台的 navItems（含子菜单）
  const navItems = currentPlatform.value?.config?.navItems || []
  const matched = findMatchingPath(navItems, route.path)
  return matched || route.path
})

/** 递归查找匹配当前路由路径的菜单项 path */
function findMatchingPath(items: any[], path: string): string | undefined {
  for (const item of items) {
    if (item.path && path.startsWith(item.path)) {
      return item.path
    }
    if (item.children?.length) {
      const childMatch = findMatchingPath(item.children, path)
      if (childMatch) return childMatch
    }
  }
  return undefined
}

/** 角色标签 */
const roleLabel = computed(() => {
  const roleMap: Record<string, string> = {
    super_admin: '超级管理员',
    admin: '管理员',
    readonly: '只读用户'
  }
  return roleMap[authStore.user?.role || ''] || authStore.user?.role || ''
})

/** 切换平台 */
function onPlatformChange(id: string) {
  platformStore.selectPlatform(id)
  const target = platformStore.getPlatformById(id)
  // 递归查找第一个有 path 的菜单项（支持二级菜单）
  const firstPath = findFirstPath(target?.config?.navItems || [])
  if (firstPath) {
    router.push(firstPath)
  } else {
    router.push('/dashboard')
  }
}

/** 递归查找第一个有 path 的菜单项 */
function findFirstPath(items: any[]): string | undefined {
  for (const item of items) {
    if (item.path) return item.path
    if (item.children?.length) {
      const childPath = findFirstPath(item.children)
      if (childPath) return childPath
    }
  }
  return undefined
}

/** 菜单选择 */
function onMenuSelect(index: string) {
  router.push(index)
}

/** 下拉命令处理 */
function handleCommand(cmd: string) {
  if (cmd === 'logout') {
    authStore.logout()
    platformStore.reset()
    router.push('/login')
  }
}

/** 监听 currentPlatform 变化，同步 select 值 */
watch(
  () => currentPlatform.value?.id,
  (newId) => {
    if (newId) {
      currentPlatformId.value = newId
    }
  },
  { immediate: true }
)

/** 监听路由变化，自动同步当前平台（如直接访问 /aluminum/overview） */
watch(
  () => route.path,
  (path) => {
    if (path === '/dashboard' || path === '/platform-management' || path === '/permission-management' || path === '/login') return
    // 从路径中提取平台ID
    const segments = path.replace(/^\//, '').split('/')
    const platformId = segments[0]
    if (platformId && platformStore.platforms.length > 0) {
      const target = platformStore.getPlatformById(platformId)
      if (target && target.id !== currentPlatform.value?.id) {
        platformStore.selectPlatform(platformId)
      }
    }
  },
  { immediate: true }
)
</script>

<style scoped lang="scss">
.main-layout {
  display: flex;
  min-height: 100vh;
  background: #0d1117;
}

/* ===== 左侧导航栏 ===== */
.sidebar {
  width: 220px;
  background: #161b22;
  border-right: 1px solid #30363d;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  position: sticky;
  top: 0;
  height: 100vh;
  overflow-y: auto;

  .platform-switcher {
    padding: 16px 12px;
    border-bottom: 1px solid #21262d;

    .switcher-icon {
      font-size: 16px;
    }
  }

  .sidebar-menu {
    border-right: none;
    background: transparent;
    flex: 1;

    :deep(.el-menu-item-group__title) {
      font-size: 11px;
      color: #7d8590;
      padding: 12px 20px 6px;
      text-transform: uppercase;
      letter-spacing: 1px;
    }

    /* 二级菜单标题（如"告警与历史"）—— 深色主题适配 */
    :deep(.el-sub-menu__title) {
      color: #c9d1d9;

      &:hover {
        background: rgba(75, 63, 227, 0.08);
        color: #e6edf3;
      }
    }

    /* 二级菜单展开后的子项 */
    :deep(.el-sub-menu .el-menu-item) {
      color: #8b949e;

      &:hover {
        background: rgba(75, 63, 227, 0.08);
        color: #e6edf3;
      }

      &.is-active {
        background: rgba(75, 63, 227, 0.15);
        color: #7c6ff5;
      }
    }

    :deep(.el-menu-item) {
      color: #c9d1d9;
      height: 44px;
      line-height: 44px;

      &:hover {
        background: rgba(75, 63, 227, 0.08);
        color: #e6edf3;
      }

      &.is-active {
        background: rgba(75, 63, 227, 0.15);
        color: #7c6ff5;
        border-right: 2px solid #4b3fe3;
      }
    }
  }
}

/* ===== 主区域 ===== */
.main-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.app-header {
  height: 56px;
  background: #161b22;
  border-bottom: 1px solid #30363d;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  position: sticky;
  top: 0;
  z-index: 100;

  .header-left {
    .current-platform-name {
      font-size: 15px;
      font-weight: 600;
      color: #e6edf3;
    }
  }

  .header-right {
    .user-info {
      display: flex;
      align-items: center;
      gap: 10px;
      cursor: pointer;
      padding: 6px 12px;
      border-radius: 8px;
      transition: background 0.2s;

      &:hover {
        background: rgba(255, 255, 255, 0.05);
      }

      .avatar {
        width: 32px;
        height: 32px;
        border-radius: 50%;
        background: linear-gradient(135deg, #4b3fe3, #7c6ff5);
        color: #fff;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 13px;
        font-weight: 700;
      }

      .username {
        font-size: 13px;
        color: #e6edf3;
        font-weight: 500;
      }
    }
  }
}

.page-content {
  flex: 1;
  padding: 20px 24px 32px;
  overflow-y: auto;
}
</style>
