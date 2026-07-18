<template>
  <div class="event-page">
    <div class="filter-bar">
      <el-date-picker v-model="startDate" type="date" size="small" value-format="YYYY-MM-DD" placeholder="开始日期" style="width: 140px" />
      <span class="sep">~</span>
      <el-date-picker v-model="endDate" type="date" size="small" value-format="YYYY-MM-DD" placeholder="结束日期" style="width: 140px" />
      <el-select v-model="filterType" placeholder="事件类型" size="small" clearable style="width: 140px">
        <el-option label="过电压" value="overvoltage" />
        <el-option label="过电压恢复" value="overvoltage_recover" />
        <el-option label="低电压" value="undervoltage" />
        <el-option label="低电压恢复" value="undervoltage_recover" />
        <el-option label="本地调控" value="local_control" />
        <el-option label="上线通知" value="online" />
      </el-select>
      <el-select v-model="filterStatus" placeholder="状态" size="small" clearable style="width: 120px">
        <el-option label="未恢复" value="active" />
        <el-option label="已恢复" value="recovered" />
      </el-select>
      <el-input v-model="filterDevice" placeholder="设备ID" size="small" clearable style="width: 160px" />
      <el-button size="small" type="primary" @click="loadEvents">查询</el-button>
      <el-button size="small" plain @click="resetFilters">重置</el-button>
    </div>

    <el-table :data="events" stripe size="small" table-layout="auto" v-loading="loading">
      <el-table-column prop="occurredAt" label="时间" width="165">
        <template #default="{ row }">{{ formatDT(row.occurredAt) }}</template>
      </el-table-column>
      <el-table-column prop="deviceId" label="设备ID" />
      <el-table-column prop="eventType" label="事件类型">
        <template #default="{ row }">
          <el-tag :type="getEventTagType(row.eventType)" size="small">{{ EVENT_TYPE_LABELS[row.eventType] || row.eventType }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="detail" label="越限详情">
        <template #default="{ row }">{{ formatDetail(row.detail) }}</template>
      </el-table-column>
      <el-table-column prop="status" label="状态">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'danger' : 'success'" size="small">
            {{ row.status === 'active' ? '未恢复' : '已恢复' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="goDevice(row.deviceId)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-pagination
        v-model:current-page="page"
        :page-size="20"
        :total="total"
        layout="total, prev, pager, next"
        small
        @current-change="loadEvents"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getEvents } from '@/api/event'
import { useWebSocket } from '@/composables/useWebSocket'
import { EVENT_TYPE_LABELS } from '@/utils/constants'

const router = useRouter()
const { on: wsOn } = useWebSocket()
const events = ref<any[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const startDate = ref('')
const endDate = ref('')
const filterType = ref('')
const filterStatus = ref('')
const filterDevice = ref('')

function formatDT(dt: string) {
  if (!dt) return '-'
  return new Date(dt).toLocaleString('zh-CN')
}

function formatDetail(detail: any) {
  if (!detail) return '-'
  try {
    if (typeof detail === 'string') detail = JSON.parse(detail)
    const parts: string[] = []
    if (detail.data) {
      Object.entries(detail.data).forEach(([k, v]) => {
        parts.push(`${k}: ${v}`)
      })
    }
    if (detail.cycle) parts.push(`周期: ${detail.cycle}`)
    if (detail.voltageHigh) parts.push(`高压: ${detail.voltageHigh}V`)
    if (detail.voltageLow) parts.push(`低压: ${detail.voltageLow}V`)
    return parts.length ? parts.join(', ') : JSON.stringify(detail)
  } catch {
    return String(detail)
  }
}

function getEventTagType(type: string) {
  if (type.includes('over')) return 'danger'
  if (type.includes('under')) return 'warning'
  return 'info'
}

function goDevice(id: string) {
  router.push(`/devices/${id}`)
}

function resetFilters() {
  startDate.value = ''
  endDate.value = ''
  filterType.value = ''
  filterStatus.value = ''
  filterDevice.value = ''
  page.value = 1
  loadEvents()
}

async function loadEvents() {
  loading.value = true
  try {
    const params: any = { page: page.value, pageSize: 20 }
    if (startDate.value) params.startDate = startDate.value
    if (endDate.value) params.endDate = endDate.value
    if (filterType.value) params.eventType = filterType.value
    if (filterStatus.value) params.status = filterStatus.value
    if (filterDevice.value) params.deviceId = filterDevice.value
    const res: any = await getEvents(params)
    events.value = res.data?.items || []
    total.value = res.data?.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(loadEvents)
</script>

<style scoped lang="scss">
.event-page {
  color: #e6ecf5;
}

.filter-bar {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;

  .sep {
    color: #566880;
    font-size: 13px;
  }
}

.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
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
