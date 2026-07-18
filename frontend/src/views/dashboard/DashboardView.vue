<template>
  <div class="dashboard">
    <el-alert
      v-if="overview?.unresolvedEvents"
      :title="`未恢复告警: ${overview.unresolvedEvents}条`"
      type="error"
      show-icon
      :closable="false"
      class="alert-bar"
    />

    <div class="stat-row">
      <div class="stat-card blue">
        <div class="stat-icon">📊</div>
        <div class="stat-info">
          <div class="stat-label">设备总数</div>
          <div class="stat-value">{{ overview?.totalDevices || 0 }}</div>
          <div class="stat-sub">在线 <span class="green">{{ overview?.onlineDevices || 0 }}</span> / 离线 <span class="red">{{ overview?.offlineDevices || 0 }}</span></div>
        </div>
      </div>
      <div class="stat-card red">
        <div class="stat-icon">🔔</div>
        <div class="stat-info">
          <div class="stat-label">今日告警数</div>
          <div class="stat-value">{{ overview?.todayEvents || 0 }}</div>
        </div>
      </div>
      <div class="stat-card orange">
        <div class="stat-icon">⚡</div>
        <div class="stat-info">
          <div class="stat-label">台区总数</div>
          <div class="stat-value">{{ overview?.totalStations || 0 }}</div>
        </div>
      </div>
      <div class="stat-card teal">
        <div class="stat-icon">🏗️</div>
        <div class="stat-info">
          <div class="stat-label">未恢复事件</div>
          <div class="stat-value">{{ overview?.unresolvedEvents || 0 }}</div>
        </div>
      </div>
    </div>

    <div class="row2">
      <div class="panel">
        <div class="panel-header">设备状态分布（按台区）</div>
        <div class="panel-body">
          <div v-for="item in deviceStatus" :key="item.stationName" class="status-row">
            <span class="status-label">{{ item.stationName }}</span>
            <el-progress
              :percentage="item.total > 0 ? Math.round(item.online / item.total * 100) : 0"
              :stroke-width="18"
              :color="getProgressColor(item.online, item.total)"
            />
            <span class="status-count" :style="{ color: getStatusColor(item.online, item.total) }">
              {{ item.online }}/{{ item.total }}
            </span>
          </div>
          <el-empty v-if="!deviceStatus?.length" description="暂无数据" />
        </div>
      </div>
      <div class="panel">
        <div class="panel-header">今日过电压 Top10</div>
        <div class="panel-body">
          <el-table :data="overvoltageTop" size="small" stripe table-layout="auto">
            <el-table-column type="index" label="#" />
            <el-table-column prop="deviceId" label="设备ID" />
            <el-table-column prop="stationName" label="台区" />
            <el-table-column prop="count" label="次数" align="right">
              <template #default="{ row }">
                <span style="color: #f6565c; font-weight: 700">{{ row.count }}次</span>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!overvoltageTop?.length" description="今日无过电压事件" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { getOverview, getDeviceStatus, getOvervoltageTop } from '@/api/stats'
import { useWebSocket } from '@/composables/useWebSocket'
import { useDeviceStore } from '@/stores/device'

const overview = ref<any>(null)
const deviceStatus = ref<any[]>([])
const overvoltageTop = ref<any[]>([])
const deviceStore = useDeviceStore()

const { on, off } = useWebSocket()

async function handleDeviceStatus(data: any) {
  // 设备状态变化时重新拉取概览，避免手动计数偏差
  try {
    const res: any = await getOverview()
    if (res.data) overview.value = res.data
  } catch {}
  // 同步更新设备树中的状态
  deviceStore.updateDeviceStatus(data.deviceId, data.status)
}

function handleRealtimeData(data: any) {
  // 实时数据推送（DC/AC），前端设备详情页可展示
  if (data.ac || data.dc) {
    deviceStore.updateRealtimeData({ ac: data.ac || null, dc: data.dc || null, cs: data.cs || 0 })
  }
}

async function handleEvent(data: any) {
  // 有新事件时重新拉取概览
  try {
    const res: any = await getOverview()
    if (res.data) overview.value = res.data
  } catch {}
}

onMounted(async () => {
  try {
    const [overviewRes, statusRes, topRes] = await Promise.all([
      getOverview() as any,
      getDeviceStatus() as any,
      getOvervoltageTop() as any
    ])
    overview.value = overviewRes.data
    deviceStatus.value = statusRes.data || []
    overvoltageTop.value = topRes.data || []
  } catch (e) {
    console.error('Dashboard load error:', e)
  }

  on('device_status', handleDeviceStatus)
  on('realtime_data', handleRealtimeData)
  on('event', handleEvent)
})

onUnmounted(() => {
  off('device_status', handleDeviceStatus)
  off('realtime_data', handleRealtimeData)
  off('event', handleEvent)
})

function getProgressColor(online: number, total: number) {
  const pct = total > 0 ? online / total : 0
  if (pct >= 0.8) return '#3dd68c'
  if (pct >= 0.5) return '#f0a030'
  return '#f6565c'
}

function getStatusColor(online: number, total: number) {
  const pct = total > 0 ? online / total : 0
  if (pct >= 0.8) return '#3dd68c'
  if (pct >= 0.5) return '#f0a030'
  return '#f6565c'
}
</script>

<style scoped lang="scss">
.dashboard {
  color: #e6ecf5;
}

.alert-bar {
  margin-bottom: 16px;
  background: rgba(246, 86, 92, 0.12);
  border: 1px solid rgba(246, 86, 92, 0.3);
  border-radius: 8px;
  color: #f6565c;
}

.stat-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}

.stat-card {
  background: #182434;
  border: 1px solid #253650;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all 0.22s ease;
  cursor: pointer;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 24px rgba(0, 0, 0, 0.4);
  }

  .stat-icon {
    font-size: 22px;
  }

  .stat-label {
    font-size: 10px;
    color: #566880;
    text-transform: uppercase;
    letter-spacing: 1px;
  }

  .stat-value {
    font-size: 28px;
    font-weight: 700;
    line-height: 1.1;
  }

  .stat-sub {
    font-size: 11px;
    color: #566880;
    margin-top: 3px;
  }

  &.blue .stat-value { color: #4da3ff; }
  &.red .stat-value { color: #f6565c; }
  &.orange .stat-value { color: #f0a030; }
  &.teal .stat-value { color: #3dd6c8; }
}

.green { color: #3dd68c; }
.red { color: #f6565c; }

.row2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.panel {
  background: #182434;
  border: 1px solid #253650;
  border-radius: 12px;
  overflow: hidden;

  .panel-header {
    padding: 14px 18px;
    border-bottom: 1px solid #253650;
    font-size: 13px;
    font-weight: 600;
    background: #1d2b3e;
  }

  .panel-body {
    padding: 16px 18px;
  }
}

.status-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;

  .status-label {
    width: 100px;
    flex-shrink: 0;
    text-align: right;
    color: #8d9db8;
    font-size: 12px;
  }

  :deep(.el-progress) {
    flex: 1;
  }

  .status-count {
    width: 60px;
    text-align: right;
    font-size: 12px;
    font-weight: 600;
  }
}
</style>
