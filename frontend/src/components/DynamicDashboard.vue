<template>
  <div class="dynamic-dashboard" v-loading="loading">
    <!-- KPI 卡片 -->
    <div class="kpi-row">
      <div class="al-card kpi-card blue">
        <div class="kpi-icon">
          <el-icon><Monitor /></el-icon>
        </div>
        <div class="kpi-info">
          <div class="kpi-label">设备总数</div>
          <div class="kpi-value">{{ stats.totalDevices }}</div>
        </div>
      </div>

      <div class="al-card kpi-card green">
        <div class="kpi-icon">
          <el-icon><CircleCheck /></el-icon>
        </div>
        <div class="kpi-info">
          <div class="kpi-label">在线设备</div>
          <div class="kpi-value">{{ stats.onlineDevices }}</div>
        </div>
      </div>

      <div class="al-card kpi-card gray">
        <div class="kpi-icon">
          <el-icon><CircleClose /></el-icon>
        </div>
        <div class="kpi-info">
          <div class="kpi-label">离线设备</div>
          <div class="kpi-value">{{ stats.offlineDevices }}</div>
        </div>
      </div>

      <div class="al-card kpi-card red">
        <div class="kpi-icon">
          <el-icon><Bell /></el-icon>
        </div>
        <div class="kpi-info">
          <div class="kpi-label">告警数量</div>
          <div class="kpi-value">{{ stats.alarmCount }}</div>
        </div>
      </div>
    </div>

    <!-- 最近告警表格 -->
    <div class="al-card alarm-panel">
      <div class="panel-header">
        <span class="panel-title">最近告警</span>
        <el-button text type="primary" size="small" @click="fetchData">刷新</el-button>
      </div>
      <el-table :data="recentAlarms" stripe size="small" max-height="400">
        <el-table-column type="index" label="#" width="50" align="center" />
        <el-table-column prop="deviceName" label="设备名称" min-width="140" show-overflow-tooltip />
        <el-table-column prop="alarmType" label="告警类型" width="120">
          <template #default="{ row }">
            <el-tag :type="getAlarmTagType(row.level)" size="small" effect="dark">
              {{ row.alarmType || row.type || '-' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="告警描述" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.message || row.detail || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="level" label="级别" width="80" align="center">
          <template #default="{ row }">
            <span :style="{ color: getLevelColor(row.level), fontWeight: 600 }">
              {{ getLevelText(row.level) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="timestamp" label="时间" width="170">
          <template #default="{ row }">
            {{ formatTime(row.occurredAt || row.timestamp || row.createdAt || row.time) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.status === 'resolved' ? 'success' : 'danger'"
              size="small"
              effect="dark"
            >
              {{ row.status === 'resolved' ? '已恢复' : '未恢复' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && recentAlarms.length === 0" description="暂无告警" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import request from '@/api/request'

const props = defineProps<{
  platform: string
  config?: any
}>()

const loading = ref(false)

const stats = reactive({
  totalDevices: 0,
  onlineDevices: 0,
  offlineDevices: 0,
  alarmCount: 0
})

const recentAlarms = ref<any[]>([])

/** 获取仪表盘数据 */
async function fetchData() {
  if (!props.platform) return
  loading.value = true
  try {
    const res = await request.get(`/${props.platform}/dashboard`, {
      skipErrorMessage: true
    })
    const data = res.data || {}
    const kpi = data.kpi || {}

    stats.totalDevices = kpi.deviceTotal ?? data.totalDevices ?? 0
    stats.onlineDevices = kpi.online ?? data.onlineDevices ?? 0
    stats.offlineDevices = kpi.offline ?? data.offlineDevices ?? 0
    stats.alarmCount = kpi.activeAlarm ?? data.alarmCount ?? 0

    recentAlarms.value = data.recentAlarms || data.alarms || []
  } catch {
    // 接口不可用时保持默认值
  } finally {
    loading.value = false
  }
}

/** 告警标签类型 */
function getAlarmTagType(level: any): 'danger' | 'warning' | 'info' {
  if (level === 1 || level === 'critical') return 'danger'
  if (level === 2 || level === 'warning') return 'warning'
  return 'info'
}

/** 级别颜色 */
function getLevelColor(level: any): string {
  if (level === 1 || level === 'critical') return '#f85149'
  if (level === 2 || level === 'warning') return '#d29922'
  return '#6e7681'
}

/** 级别文本 */
function getLevelText(level: any): string {
  if (level === 1 || level === 'critical') return '紧急'
  if (level === 2 || level === 'warning') return '警告'
  if (level === 3 || level === 'info') return '提示'
  return String(level || '-')
}

/** 格式化时间 */
function formatTime(ts: any): string {
  if (!ts) return '-'
  const d = new Date(ts)
  if (isNaN(d.getTime())) return String(ts)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

onMounted(() => {
  fetchData()
})

defineExpose({ fetchData })
</script>

<style scoped lang="scss">
.dynamic-dashboard {
  .kpi-row {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
    margin-bottom: 20px;
  }

  .kpi-card {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 20px;

    .kpi-icon {
      width: 48px;
      height: 48px;
      border-radius: 10px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 24px;
    }

    .kpi-info {
      flex: 1;

      .kpi-label {
        font-size: 12px;
        color: #8b949e;
        margin-bottom: 4px;
      }

      .kpi-value {
        font-size: 28px;
        font-weight: 700;
        color: #e6edf3;
        font-variant-numeric: tabular-nums;
      }
    }

    &.blue .kpi-icon {
      background: rgba(88, 166, 255, 0.12);
      color: #58a6ff;
    }

    &.green .kpi-icon {
      background: rgba(63, 185, 80, 0.12);
      color: #3fb950;
    }

    &.gray .kpi-icon {
      background: rgba(110, 118, 129, 0.12);
      color: #6e7681;
    }

    &.red .kpi-icon {
      background: rgba(248, 81, 73, 0.12);
      color: #f85149;
    }
  }

  .alarm-panel {
    .panel-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 16px 20px;
      border-bottom: 1px solid #21262d;

      .panel-title {
        font-size: 15px;
        font-weight: 600;
        color: #e6edf3;
      }
    }
  }
}
</style>
