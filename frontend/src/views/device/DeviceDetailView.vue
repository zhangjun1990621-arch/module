<template>
  <div class="device-detail">
    <div class="device-layout">
      <div class="device-sidebar">
        <div class="search-area">
          <el-input v-model="searchKeyword" placeholder="搜索设备ID..." size="small" clearable />
          <el-select v-model="filterStation" placeholder="全部台区" size="small" clearable style="width: 100%; margin-top: 8px">
            <el-option v-for="st in deviceStore.deviceTree" :key="st.id" :label="st.name" :value="st.id" />
          </el-select>
        </div>
        <div class="device-list">
          <div v-for="station in filteredTree" :key="station.id" class="station-group">
            <div class="group-header">
              <el-icon><ArrowDown /></el-icon>
              {{ station.name }}
              <span class="count" :style="{ color: getStationColor(station) }">
                {{ getOnlineCount(station) }}/{{ station.devices?.length || 0 }}
              </span>
            </div>
            <div
              v-for="device in station.devices"
              :key="device.id"
              class="device-item"
              :class="{ active: deviceStore.selectedDeviceId === device.id }"
              @click="selectDevice(device.id)"
            >
              <span class="dot" :class="device.status === 'online' ? 'live' : 'dead'"></span>
              {{ device.id }}
            </div>
          </div>
          <el-empty v-if="!deviceStore.deviceTree?.length" description="暂无设备" />
        </div>
      </div>

      <div class="device-right">
        <template v-if="deviceStore.selectedDevice">
          <div class="device-info-bar">
            <span class="device-id">{{ deviceStore.selectedDevice.id }}</span>
            <span class="device-meta">型号: {{ deviceStore.selectedDevice.model || '-' }}</span>
            <span class="device-meta">SN: {{ deviceStore.selectedDevice.serialNo || '-' }}</span>
            <span class="device-meta">容量: {{ deviceStore.selectedDevice.capacity || '-' }}W</span>
            <span class="device-meta">软件: {{ deviceStore.selectedDevice.software || '-' }}</span>
            <span class="device-meta">硬件: {{ deviceStore.selectedDevice.hardware || '-' }}</span>
            <span class="device-meta">
              信号:
              <span :class="signalClass">{{ deviceStore.selectedDevice.signalStrength ?? '-' }}</span>
            </span>
            <span class="device-meta">{{ lastOnlineText }}</span>
            <el-tag :type="deviceStore.selectedDevice.status === 'online' ? 'success' : 'info'" size="small">
              {{ deviceStore.selectedDevice.status === 'online' ? '在线' : '离线' }}
            </el-tag>
            <el-button size="small" type="primary" plain @click="refreshDevice" style="margin-left: auto">刷新</el-button>
          </div>

          <div class="realtime-panel">
            <div class="panel-header">
              实时数据
              <span v-if="lastUpdate" class="update-time">更新于 {{ lastUpdate }}</span>
            </div>
            <div class="rt-grid">
              <div class="rt-cell" v-for="item in rtItems" :key="item.label">
                <div class="rt-label">{{ item.label }}</div>
                <div class="rt-value">{{ item.value }}<span class="rt-unit">{{ item.unit }}</span></div>
              </div>
            </div>
          </div>

          <div class="action-bar">
            <el-button type="primary" :loading="pollingLoading" @click="handlePoll">召测</el-button>
            <el-button type="primary" plain @click="showSetting = true">设置参数</el-button>
            <el-button type="info" plain :loading="ackLoading" @click="handleReportAck">OTA准备</el-button>
            <el-button type="warning" plain :loading="rebootLoading" @click="handleReboot">重启</el-button>
            <el-button type="danger" plain :loading="factoryLoading" @click="handleFactory">恢复出厂</el-button>
          </div>

          <div class="curve-panel">
            <div class="panel-header">电压 / 电流曲线</div>
            <div class="curve-row">
              <div class="curve-box">
                <div class="curve-title">电压曲线</div>
                <div ref="voltageChartEl" class="chart-el"></div>
              </div>
              <div class="curve-box">
                <div class="curve-title">电流曲线</div>
                <div ref="currentChartEl" class="chart-el"></div>
              </div>
            </div>
          </div>

          <el-tabs v-model="activeTab">
            <el-tab-pane label="历史数据" name="history">
              <div class="filter-bar">
                <el-date-picker
                  v-model="dateRange"
                  type="daterange"
                  size="small"
                  value-format="YYYY-MM-DD"
                  start-placeholder="开始"
                  end-placeholder="结束"
                  :clearable="false"
                  class="compact-date-picker"
                />
                <el-button size="small" type="primary" @click="loadHistory">查询</el-button>
                <el-button size="small" plain @click="handleExport">导出CSV</el-button>
              </div>
              <el-table :data="historyData" size="small" stripe max-height="400">
                <el-table-column prop="timestamp" label="时间" width="100" />
                <el-table-column prop="phaseAV" label="A相(V)" width="80" />
                <el-table-column prop="phaseBV" label="B相(V)" width="80" />
                <el-table-column prop="phaseCV" label="C相(V)" width="80" />
                <el-table-column prop="activePower" label="有功(W)" width="100" />
                <el-table-column prop="reactivePower" label="无功(Var)" width="100" />
                <el-table-column prop="powerFactor" label="PF" width="60" />
                <el-table-column prop="frequency" label="频率(Hz)" width="80" />
              </el-table>
            </el-tab-pane>
            <el-tab-pane label="通信日志" name="comm">
              <el-table :data="mqttLogs" size="small" stripe max-height="400" v-loading="mqttLogLoading">
                <el-table-column prop="createdAt" label="时间" width="170">
                  <template #default="{ row }">{{ new Date(row.createdAt).toLocaleString('zh-CN') }}</template>
                </el-table-column>
                <el-table-column prop="direction" label="方向" width="80">
                  <template #default="{ row }">
                    <el-tag :type="row.direction === 'up' ? 'success' : 'primary'" size="small">
                      {{ row.direction === 'up' ? '上行' : '下行' }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="topic" label="Topic" width="120" />
                <el-table-column prop="payload" label="内容" min-width="200">
                  <template #default="{ row }">{{ formatPayload(row.payload) }}</template>
                </el-table-column>
              </el-table>
            </el-tab-pane>
            <el-tab-pane label="操作日志" name="oplog">
              <el-table :data="opLogs" size="small" stripe max-height="400" v-loading="opLogLoading">
                <el-table-column prop="createdAt" label="时间" width="170">
                  <template #default="{ row }">{{ new Date(row.createdAt).toLocaleString('zh-CN') }}</template>
                </el-table-column>
                <el-table-column prop="operator" label="操作人" width="100" />
                <el-table-column prop="actionType" label="操作类型" width="100" />
                <el-table-column prop="content" label="内容" min-width="200" />
                <el-table-column prop="result" label="结果" width="80">
                  <template #default="{ row }">
                    <el-tag :type="row.result === 'success' ? 'success' : 'danger'" size="small">
                      {{ row.result === 'success' ? '成功' : '失败' }}
                    </el-tag>
                  </template>
                </el-table-column>
              </el-table>
            </el-tab-pane>
          </el-tabs>
        </template>
        <el-empty v-else description="请选择一个设备" />
      </div>
    </div>

    <el-dialog v-model="showSetting" title="设置参数" width="500px">
      <el-form label-width="120px">
        <el-form-item label="过压保护(V)">
          <el-input-number v-model="settingForm.overVoltage" :min="220" :max="280" :step="1" />
        </el-form-item>
        <el-form-item label="欠压保护(V)">
          <el-input-number v-model="settingForm.underVoltage" :min="170" :max="220" :step="1" />
        </el-form-item>
        <el-form-item label="有功功率(%)">
          <el-slider v-model="settingForm.activePower" :min="0" :max="100" show-input />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showSetting = false">取消</el-button>
        <el-button type="primary" @click="handleSet">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDeviceStore } from '@/stores/device'
import { ArrowDown } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { getHistoryData } from '@/api/data'
import { getOperationLogs, getMQTTLogs } from '@/api/log'
import request from '@/api/pvRequest'
import { useCurveChart } from '@/composables/useCurveChart'
import { useWebSocket } from '@/composables/useWebSocket'

const route = useRoute()
const router = useRouter()
const deviceStore = useDeviceStore()

const searchKeyword = ref('')
const filterStation = ref<number | null>(null)
const activeTab = ref('history')
const dateRange = ref<[string, string]>(['', ''])
const historyData = ref<any[]>([])
const mqttLogs = ref<any[]>([])
const opLogs = ref<any[]>([])
const lastUpdate = ref('')
const showSetting = ref(false)
const pollingLoading = ref(false)
const rebootLoading = ref(false)
const factoryLoading = ref(false)
const ackLoading = ref(false)
const mqttLogLoading = ref(false)
const opLogLoading = ref(false)

const { on: onWS, off: offWS } = useWebSocket()

const voltageChartEl = ref<HTMLElement | null>(null)
const currentChartEl = ref<HTMLElement | null>(null)

const voltageChart = useCurveChart(voltageChartEl, ['A相', 'B相', 'C相'], ['#4da3ff', '#3dd68c', '#f0a030'], 'V')
const currentChart = useCurveChart(currentChartEl, ['A相', 'B相', 'C相'], ['#4da3ff', '#3dd68c', '#f0a030'], 'A')

let realtimeUnwatch: (() => void) | null = null

const settingForm = ref({
  overVoltage: 265,
  underVoltage: 195,
  activePower: 100
})

const filteredTree = computed(() => {
  let tree = deviceStore.deviceTree || []
  if (filterStation.value !== null && filterStation.value !== undefined) {
    const stationId = Number(filterStation.value)
    tree = tree.filter((s: any) => s.id === stationId)
  }
  if (searchKeyword.value) {
    tree = tree.map((s: any) => ({
      ...s,
      devices: (s.devices || []).filter((d: any) => d.id.includes(searchKeyword.value))
    })).filter((s: any) => s.devices.length > 0)
  }
  return tree
})

const rtItems = computed(() => {
  const rt = deviceStore.realtimeData
  if (rt?.ac) {
    return [
      { label: 'A相电压', value: rt.ac.v?.[0] || '-', unit: 'V' },
      { label: 'B相电压', value: rt.ac.v?.[1] || '-', unit: 'V' },
      { label: 'C相电压', value: rt.ac.v?.[2] || '-', unit: 'V' },
      { label: '有功功率', value: rt.ac.p || '-', unit: 'W' },
      { label: '功率因数', value: rt.ac.pf || '-', unit: '' },
      { label: 'A相电流', value: rt.ac.c?.[0] || '-', unit: 'A' },
      { label: 'B相电流', value: rt.ac.c?.[1] || '-', unit: 'A' },
      { label: 'C相电流', value: rt.ac.c?.[2] || '-', unit: 'A' },
      { label: '无功功率', value: rt.ac.q || '-', unit: 'Var' },
      { label: '电网频率', value: rt.ac.f || '-', unit: 'Hz' }
    ]
  }
  if (rt?.dc) {
    return [
      { label: 'PV1电压', value: rt.dc.v?.[0] || '-', unit: 'V' },
      { label: 'PV2电压', value: rt.dc.v?.[1] || '-', unit: 'V' },
      { label: 'PV3电压', value: rt.dc.v?.[2] || '-', unit: 'V' },
      { label: 'PV1电流', value: rt.dc.c?.[0] || '-', unit: 'A' },
      { label: 'PV2电流', value: rt.dc.c?.[1] || '-', unit: 'A' },
      { label: 'PV3电流', value: rt.dc.c?.[2] || '-', unit: 'A' },
      { label: 'PV1功率', value: rt.dc.p?.[0] || '-', unit: 'W' },
      { label: 'PV2功率', value: rt.dc.p?.[1] || '-', unit: 'W' },
      { label: 'PV3功率', value: rt.dc.p?.[2] || '-', unit: 'W' },
      { label: 'DC路数', value: rt.dc.v?.length || '-', unit: '路' }
    ]
  }
  return [
    { label: 'A相电压', value: '-', unit: 'V' },
    { label: 'B相电压', value: '-', unit: 'V' },
    { label: 'C相电压', value: '-', unit: 'V' },
    { label: '有功功率', value: '-', unit: 'W' },
    { label: '功率因数', value: '-', unit: '' },
    { label: 'A相电流', value: '-', unit: 'A' },
    { label: 'B相电流', value: '-', unit: 'A' },
    { label: 'C相电流', value: '-', unit: 'A' },
    { label: '无功功率', value: '-', unit: 'Var' },
    { label: '电网频率', value: '-', unit: 'Hz' }
  ]
})

const signalClass = computed(() => {
  const s = deviceStore.selectedDevice?.signalStrength
  if (s == null) return ''
  if (s >= 25) return 'signal-strong'
  if (s >= 15) return 'signal-mid'
  return 'signal-weak'
})

const lastOnlineText = computed(() => {
  const t = deviceStore.selectedDevice?.lastOnline
  if (!t) return ''
  return '最后在线: ' + new Date(t).toLocaleString('zh-CN')
})

function getOnlineCount(station: any) {
  return (station.devices || []).filter((d: any) => d.status === 'online').length
}

function getStationColor(station: any) {
  const total = station.devices?.length || 0
  const online = getOnlineCount(station)
  const pct = total > 0 ? online / total : 0
  if (pct >= 0.8) return '#3dd68c'
  if (pct >= 0.5) return '#f0a030'
  return '#f6565c'
}

async function selectDevice(id: string) {
  await deviceStore.selectDevice(id)
  lastUpdate.value = new Date().toLocaleTimeString()
  loadDeviceLogs(id)
}

function refreshDevice() {
  if (deviceStore.selectedDeviceId) {
    selectDevice(deviceStore.selectedDeviceId)
  }
}

async function handlePoll() {
  if (!deviceStore.selectedDeviceId) return
  pollingLoading.value = true
  try {
    await request.post(`/devices/${deviceStore.selectedDeviceId}/polling`, { items: ['ac', 'dc', 'sw', 'hw'] })
    ElMessage.success('召测指令已下发')
    setTimeout(refreshDevice, 2000)
  } catch (e: any) {
    ElMessage.error(e?.message || '召测失败')
  } finally {
    pollingLoading.value = false
  }
}

async function handleReboot() {
  if (!deviceStore.selectedDeviceId) return
  await ElMessageBox.confirm('确认重启设备？', '提示', { type: 'warning' })
  rebootLoading.value = true
  try {
    await request.post(`/devices/${deviceStore.selectedDeviceId}/reboot`)
    ElMessage.success('重启指令已下发')
  } catch (e: any) {
    ElMessage.error(e?.message || '重启失败')
  } finally {
    rebootLoading.value = false
  }
}

async function handleFactory() {
  if (!deviceStore.selectedDeviceId) return
  await ElMessageBox.confirm('确认恢复出厂设置？此操作不可恢复！', '警告', { type: 'error' })
  factoryLoading.value = true
  try {
    await request.post(`/devices/${deviceStore.selectedDeviceId}/factory`)
    ElMessage.success('恢复出厂指令已下发')
  } catch (e: any) {
    ElMessage.error(e?.message || '恢复出厂失败')
  } finally {
    factoryLoading.value = false
  }
}

async function handleReportAck() {
  if (!deviceStore.selectedDeviceId) return
  await ElMessageBox.confirm(
    '将向设备下发上报确认指令（dn/rr），设备收到后将进入OTA准备状态，停止正常上报。确认发送？',
    'OTA准备',
    { type: 'warning' }
  )
  ackLoading.value = true
  try {
    await request.post(`/devices/${deviceStore.selectedDeviceId}/report-ack`)
    ElMessage.success('OTA准备指令已下发，即将跳转升级页面')
    setTimeout(() => {
      router.push('/ota')
    }, 1000)
  } catch (e: any) {
    ElMessage.error(e?.message || '下发失败')
  } finally {
    ackLoading.value = false
  }
}

async function handleSet() {
  if (!deviceStore.selectedDeviceId) return
  try {
    await request.post(`/devices/${deviceStore.selectedDeviceId}/set`, {
      eov: settingForm.value.overVoltage,
      euv: settingForm.value.underVoltage,
      ap: settingForm.value.activePower
    })
    ElMessage.success('参数设置已下发')
    showSetting.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || '设置失败')
  }
}

async function loadHistory() {
  if (!deviceStore.selectedDeviceId) return
  try {
    const [start, end] = dateRange.value
    if (!start || !end) {
      const today = new Date().toISOString().split('T')[0]
      const weekAgo = new Date(Date.now() - 7 * 86400000).toISOString().split('T')[0]
      dateRange.value = [weekAgo, today]
      const res: any = await getHistoryData({ deviceId: deviceStore.selectedDeviceId, start: weekAgo, end: today })
      historyData.value = res.data || []
    } else {
      const res: any = await getHistoryData({ deviceId: deviceStore.selectedDeviceId, start, end })
      historyData.value = res.data || []
    }
  } catch (e) {
    console.error(e)
  }
}

function handleExport() {
  if (!deviceStore.selectedDeviceId) return
  const [start, end] = dateRange.value
  const token = localStorage.getItem('token')
  window.open(`/api/data/export?deviceId=${deviceStore.selectedDeviceId}&start=${start}&end=${end}&token=${token}`)
}

function formatPayload(payload: any) {
  if (!payload) return '-'
  try {
    if (typeof payload === 'string') payload = JSON.parse(payload)
    return JSON.stringify(payload).substring(0, 120) + '...'
  } catch {
    return String(payload).substring(0, 120)
  }
}

async function loadDeviceLogs(deviceId: string) {
  mqttLogLoading.value = true
  opLogLoading.value = true
  try {
    const [mqttRes, opRes]: any = await Promise.all([
      getMQTTLogs({ deviceId, pageSize: 20 }),
      getOperationLogs({ deviceId, pageSize: 20 })
    ])
    mqttLogs.value = mqttRes.data?.items || []
    opLogs.value = opRes.data?.items || []
  } catch (e) {
    console.error(e)
  } finally {
    mqttLogLoading.value = false
    opLogLoading.value = false
  }
}

function handleRealtimeData(data: any) {
  if (data.deviceId !== deviceStore.selectedDeviceId) return
  if (data.ac || data.dc) {
    deviceStore.updateRealtimeData({ ac: data.ac || null, dc: data.dc || null, cs: 0 })
  }
}

function handleDeviceStatus(data: any) {
  deviceStore.updateDeviceStatus(data.deviceId, data.status)
  if (data.deviceId === deviceStore.selectedDeviceId && deviceStore.selectedDevice) {
    deviceStore.selectedDevice.status = data.status
  }
}

onMounted(async () => {
  await deviceStore.fetchDeviceTree()
  if (route.params.id) {
    selectDevice(route.params.id as string)
  } else if (deviceStore.deviceTree.length > 0) {
    const firstDevice = deviceStore.deviceTree[0]?.devices?.[0]
    if (firstDevice) selectDevice(firstDevice.id)
  }
  await nextTick()
  voltageChart.initChart()
  currentChart.initChart()

  onWS('realtime_data', handleRealtimeData)
  onWS('device_status', handleDeviceStatus)

  realtimeUnwatch = watch(
    () => deviceStore.realtimeData,
    (rt) => {
      if (rt?.ac) {
        const vArr = rt.ac.v || []
        const cArr = rt.ac.c || []
        voltageChart.pushPoint([vArr[0] ?? 0, vArr[1] ?? 0, vArr[2] ?? 0])
        currentChart.pushPoint([cArr[0] ?? 0, cArr[1] ?? 0, cArr[2] ?? 0])
      } else if (rt?.dc) {
        const vArr = rt.dc.v || []
        const cArr = rt.dc.c || []
        voltageChart.pushPoint([vArr[0] ?? 0, vArr[1] ?? 0, vArr[2] ?? 0])
        currentChart.pushPoint([cArr[0] ?? 0, cArr[1] ?? 0, cArr[2] ?? 0])
      }
    },
    { deep: true }
  )
})

onBeforeUnmount(() => {
  offWS('realtime_data', handleRealtimeData)
  offWS('device_status', handleDeviceStatus)
  realtimeUnwatch?.()
  voltageChart.disposeChart()
  currentChart.disposeChart()
})
</script>

<style scoped lang="scss">
.device-layout {
  display: grid;
  grid-template-columns: 260px 1fr;
  gap: 16px;
  align-items: stretch;
}

.device-sidebar {
  background: #182434;
  border: 1px solid #253650;
  border-radius: 12px;
  overflow: hidden;
  display: flex;
  flex-direction: column;

  .search-area {
    padding: 14px;
    border-bottom: 1px solid rgba(100, 130, 180, 0.08);
    flex-shrink: 0;
  }

  .device-list {
    flex: 1;
    overflow-y: auto;
  }

  .station-group {
    .group-header {
      padding: 10px 14px;
      font-size: 12px;
      font-weight: 600;
      color: #4da3ff;
      cursor: pointer;
      display: flex;
      align-items: center;
      gap: 6px;
      border-bottom: 1px solid #253650;

      .count {
        margin-left: auto;
        font-weight: 400;
      }
    }

    .device-item {
      padding: 8px 14px 8px 30px;
      font-size: 12px;
      cursor: pointer;
      display: flex;
      align-items: center;
      gap: 8px;
      border-left: 2px solid transparent;
      transition: all 0.22s;

      &:hover { background: #1d2b3e; }
      &.active {
        background: rgba(43, 111, 212, 0.05);
        border-left-color: #4da3ff;
      }

      .dot {
        width: 7px;
        height: 7px;
        border-radius: 50%;
        &.live { background: #3dd68c; box-shadow: 0 0 8px #3dd68c; }
        &.dead { background: #4a5568; }
      }
    }
  }
}

.device-right {
  .device-info-bar {
    background: #182434;
    border: 1px solid #253650;
    border-radius: 12px;
    padding: 14px 18px;
    display: flex;
    align-items: center;
    gap: 20px;
    flex-wrap: wrap;
    margin-bottom: 12px;

    .device-id { font-size: 15px; font-weight: 700; }
    .device-meta { font-size: 12px; color: #566880; }
    .signal-strong { color: #3dd68c; font-weight: 600; }
    .signal-mid   { color: #f0a030; font-weight: 600; }
    .signal-weak  { color: #f6565c; font-weight: 600; }
  }

  .realtime-panel {
    background: #182434;
    border: 1px solid #253650;
    border-radius: 12px;
    margin-bottom: 12px;

    .panel-header {
      padding: 14px 18px;
      border-bottom: 1px solid #253650;
      font-size: 13px;
      font-weight: 600;
      background: #1d2b3e;
      display: flex;
      align-items: center;
      justify-content: space-between;

      .update-time {
        font-weight: 400;
        color: #566880;
        font-size: 11px;
      }
    }

    .rt-grid {
      display: grid;
      grid-template-columns: repeat(5, 1fr);
      gap: 10px;
      padding: 16px 18px;

      .rt-cell {
        background: #1d2b3e;
        border: 1px solid #253650;
        border-radius: 8px;
        padding: 14px;
        text-align: center;

        .rt-label {
          font-size: 10px;
          color: #566880;
          text-transform: uppercase;
          letter-spacing: 0.5px;
          margin-bottom: 6px;
        }

        .rt-value {
          font-size: 19px;
          font-weight: 700;
          color: #4da3ff;
        }

        .rt-unit {
          font-size: 11px;
          color: #566880;
          margin-left: 2px;
        }
      }
    }
  }

  .action-bar {
    display: flex;
    gap: 8px;
    margin: 14px 0;
    flex-wrap: wrap;
  }

  .filter-bar {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-bottom: 14px;
  }
}

.curve-panel {
  background: #182434;
  border: 1px solid #253650;
  border-radius: 12px;
  margin-bottom: 12px;

  .panel-header {
    padding: 12px 18px;
    border-bottom: 1px solid #253650;
    font-size: 13px;
    font-weight: 600;
    background: #1d2b3e;
  }

  .curve-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1px;
    background: #253650;
  }

  .curve-box {
    background: #182434;
    padding: 12px;

    .curve-title {
      font-size: 11px;
      color: #566880;
      margin-bottom: 4px;
      text-align: center;
    }

    .chart-el {
      width: 100%;
      height: 220px;
    }
  }
}

:deep(.el-table) {
  --el-table-bg-color: #182434;
  --el-table-tr-bg-color: #182434;
  --el-table-header-bg-color: #1d2b3e;
  --el-table-row-hover-bg-color: rgba(43, 111, 212, 0.03);
  --el-table-border-color: #253650;
  --el-table-text-color: #e6ecf5;
  --el-table-header-text-color: #566880;
}
</style>

<!-- 非 scoped 全局样式：强制缩小 daterange 选择器（Element Plus 注入内联 width，必须 !important 覆盖） -->
<style lang="scss">
.compact-date-picker.el-date-editor.el-range-editor {
  width: 220px !important;
  max-width: 220px !important;
  min-width: unset !important;
}
</style>
