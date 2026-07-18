<template>
  <div class="dashboard-view">
    <div class="page-header">
      <h2>平台概览</h2>
      <p class="header-desc">所有已注册的能源管控平台，点击卡片进入对应平台</p>
    </div>

    <div v-loading="loading" class="platform-grid">
      <div
        v-for="platform in platformStore.visiblePlatforms"
        :key="platform.id"
        class="platform-card"
        @click="enterPlatform(platform)"
      >
        <div class="card-top">
          <div class="platform-icon">{{ platform.icon }}</div>
          <div class="platform-info">
            <div class="platform-name">{{ platform.name }}</div>
            <div class="platform-id">{{ platform.id }} · {{ platform.schema }}</div>
          </div>
          <el-tag
            :type="platform.status === 'active' ? 'success' : 'info'"
            size="small"
            effect="dark"
          >
            {{ platform.status === 'active' ? '运行中' : '已停用' }}
          </el-tag>
        </div>

        <div class="card-stats">
          <div class="stat-item">
            <div class="stat-value">{{ statsMap[platform.id]?.totalDevices ?? '-' }}</div>
            <div class="stat-label">设备总数</div>
          </div>
          <div class="stat-item online">
            <div class="stat-value">{{ statsMap[platform.id]?.onlineDevices ?? '-' }}</div>
            <div class="stat-label">在线</div>
          </div>
          <div class="stat-item offline">
            <div class="stat-value">{{ statsMap[platform.id]?.offlineDevices ?? '-' }}</div>
            <div class="stat-label">离线</div>
          </div>
          <div class="stat-item alarm">
            <div class="stat-value">{{ statsMap[platform.id]?.alarmCount ?? '-' }}</div>
            <div class="stat-label">告警</div>
          </div>
        </div>

        <div class="card-footer">
          <span class="menu-count">{{ platform.config?.navItems?.length || 0 }} 个功能模块</span>
          <span class="enter-btn">进入平台 →</span>
        </div>
      </div>

      <el-empty
        v-if="!loading && platformStore.visiblePlatforms.length === 0"
        description="暂无可用平台"
        class="empty-state"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { usePlatformStore } from '@/stores/platform'
import request from '@/api/request'
import type { Platform } from '@/types/platform'

const router = useRouter()
const platformStore = usePlatformStore()
const loading = ref(false)

interface PlatformStats {
  totalDevices: number
  onlineDevices: number
  offlineDevices: number
  alarmCount: number
}

const statsMap = reactive<Record<string, PlatformStats>>({})

/** 获取各平台统计数据 */
async function loadStats(platform: Platform) {
  try {
    const res = await request.get(`/${platform.id}/dashboard`)
    const data = res.data || {}
    statsMap[platform.id] = {
      totalDevices: data.totalDevices ?? 0,
      onlineDevices: data.onlineDevices ?? 0,
      offlineDevices: data.offlineDevices ?? 0,
      alarmCount: data.alarmCount ?? data.todayAlarms ?? 0
    }
  } catch {
    // 接口不可用时不阻塞页面展示
    statsMap[platform.id] = {
      totalDevices: 0,
      onlineDevices: 0,
      offlineDevices: 0,
      alarmCount: 0
    }
  }
}

/** 进入平台 */
function enterPlatform(platform: Platform) {
  platformStore.selectPlatform(platform.id)
  const firstPath = platform.config?.navItems?.[0]?.path
  if (firstPath) {
    router.push(firstPath)
  }
}

onMounted(async () => {
  loading.value = true
  // 并行加载所有平台的统计数据
  await Promise.allSettled(
    platformStore.visiblePlatforms.map((p) => loadStats(p))
  )
  loading.value = false
})
</script>

<style scoped lang="scss">
.dashboard-view {
  .page-header {
    margin-bottom: 24px;

    h2 {
      font-size: 20px;
      font-weight: 700;
      color: #e6edf3;
    }

    .header-desc {
      font-size: 13px;
      color: #8b949e;
      margin-top: 4px;
    }
  }

  .platform-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
    gap: 16px;
    min-height: 200px;
  }

  .platform-card {
    background: #161b22;
    border: 1px solid #30363d;
    border-radius: 12px;
    padding: 20px;
    cursor: pointer;
    transition: all 0.25s ease;

    &:hover {
      border-color: #4b3fe3;
      box-shadow: 0 8px 24px rgba(75, 63, 227, 0.15);
      transform: translateY(-2px);
    }
  }

  .card-top {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 20px;

    .platform-icon {
      font-size: 32px;
      width: 48px;
      height: 48px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: rgba(75, 63, 227, 0.1);
      border-radius: 10px;
    }

    .platform-info {
      flex: 1;

      .platform-name {
        font-size: 16px;
        font-weight: 600;
        color: #e6edf3;
      }

      .platform-id {
        font-size: 12px;
        color: #6e7681;
        margin-top: 2px;
      }
    }
  }

  .card-stats {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 12px;
    padding: 16px 0;
    border-top: 1px solid #21262d;
    border-bottom: 1px solid #21262d;

    .stat-item {
      text-align: center;

      .stat-value {
        font-size: 22px;
        font-weight: 700;
        color: #e6edf3;
      }

      .stat-label {
        font-size: 11px;
        color: #6e7681;
        margin-top: 4px;
      }

      &.online .stat-value {
        color: #3fb950;
      }

      &.offline .stat-value {
        color: #6e7681;
      }

      &.alarm .stat-value {
        color: #f85149;
      }
    }
  }

  .card-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 16px;
    font-size: 12px;

    .menu-count {
      color: #6e7681;
    }

    .enter-btn {
      color: #4b3fe3;
      font-weight: 600;
    }
  }

  .empty-state {
    grid-column: 1 / -1;
    padding: 60px 0;
  }
}
</style>
