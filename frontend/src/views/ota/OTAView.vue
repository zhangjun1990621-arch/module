<template>
  <div class="ota-page">
    <div class="section">
      <div class="section-header">固件管理</div>
      <div class="section-body">
        <el-upload
          :auto-upload="false"
          :on-change="handleFileChange"
          :limit="1"
          accept=".bin,.hex,.img"
        >
          <el-button type="primary" size="small">上传固件</el-button>
        </el-upload>
        <el-table :data="firmwares" size="small" stripe table-layout="auto" style="margin-top: 12px">
          <el-table-column prop="name" label="文件名" />
          <el-table-column prop="version" label="版本" />
          <el-table-column prop="fileSize" label="大小">
            <template #default="{ row }">{{ formatSize(row.fileSize) }}</template>
          </el-table-column>
          <el-table-column prop="uploadTime" label="上传时间" width="165">
            <template #default="{ row }">{{ formatDT(row.uploadTime) }}</template>
          </el-table-column>
          <el-table-column label="操作" align="center" width="80">
            <template #default="{ row }">
              <el-button link type="danger" size="small" @click="handleDeleteFirmware(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <div class="section">
      <div class="section-header">升级任务</div>
      <div class="section-body">
        <el-button type="primary" size="small" @click="openCreateTask">创建升级任务</el-button>
        <el-table :data="tasks" size="small" stripe table-layout="auto" style="margin-top: 12px">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column label="固件">
            <template #default="{ row }">{{ row.firmware?.name || '-' }}</template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="90">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)" size="small">{{ UPGRADE_STATUS_LABELS[row.status] || row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="进度" min-width="180">
            <template #default="{ row }">
              <el-progress
                :percentage="Math.round(row.progress)"
                :stroke-width="14"
                :status="getProgressStatus(row)"
              />
              <div class="progress-detail">{{ row.successCount }}成功 / {{ row.failCount }}失败 / 共{{ row.totalDevices }}台</div>
            </template>
          </el-table-column>
          <el-table-column label="结果" width="100" v-if="hasEndedTask">
            <template #default="{ row }">
              <span v-if="row.endReason" class="end-reason" :class="row.endReason">
                {{ UPGRADE_END_REASON_LABELS[row.endReason] || row.endReason }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="创建时间" width="155">
            <template #default="{ row }">{{ formatDT(row.createdAt) }}</template>
          </el-table-column>
          <el-table-column label="操作" align="center" width="200">
            <template #default="{ row }">
              <!-- 运行中：显示暂停 -->
              <el-button v-if="row.status === 'running'" link type="warning" size="small" @click="pauseTask(row.id)">暂停</el-button>
              <!-- 已暂停：显示继续 -->
              <el-button v-if="row.status === 'paused'" link type="primary" size="small" @click="resumeTask(row.id)">继续</el-button>
              <!-- 待执行：显示开始 -->
              <el-button v-if="row.status === 'pending'" link type="primary" size="small" @click="resumeTask(row.id)">开始</el-button>
              <!-- 运行中/已暂停：显示完成（手动结束） -->
              <el-button v-if="row.status === 'running' || row.status === 'paused'" link type="success" size="small" @click="completeTask(row.id)">完成</el-button>
              <!-- 非 cancelled 且非 completed：显示取消 -->
              <el-button v-if="row.status !== 'cancelled' && row.status !== 'completed'" link type="danger" size="small" @click="cancelTask(row.id)">取消</el-button>
              <!-- 已完成且有失败设备：显示重试 -->
              <el-button v-if="row.status === 'completed' && row.failCount > 0" link type="warning" size="small" @click="retryTask(row.id)">重试</el-button>
              <!-- 非运行中/非暂停：显示删除 -->
              <el-button v-if="row.status !== 'running' && row.status !== 'paused'" link type="danger" size="small" @click="handleDeleteTask(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <el-dialog v-model="showCreateTask" title="创建升级任务" width="520px">
      <el-form :model="taskForm" label-width="80px">
        <el-form-item label="选择固件">
          <el-select v-model="taskForm.firmwareId" placeholder="请选择固件" style="width: 100%">
            <el-option
              v-for="fw in firmwares"
              :key="fw.id"
              :label="`${fw.name} (v${fw.version || '-'})`"
              :value="fw.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="选择设备">
          <div class="device-select-area">
            <div class="select-actions">
              <el-button size="small" @click="selectAllDevices">全选</el-button>
              <el-button size="small" @click="taskForm.deviceIds = []">清空</el-button>
              <span class="selected-count">已选 {{ taskForm.deviceIds.length }} 台</span>
            </div>
            <el-checkbox-group v-model="taskForm.deviceIds" class="device-checkboxes">
              <el-checkbox v-for="dev in devices" :key="dev.id" :value="dev.id">
                {{ dev.deviceId || dev.id }} - {{ dev.name }} <span class="device-status-tag" :class="dev.status">{{ dev.status === 'online' ? '在线' : '离线' }}</span>
              </el-checkbox>
            </el-checkbox-group>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateTask = false">取消</el-button>
        <el-button type="primary" @click="handleCreateTask" :loading="creatingTask">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getFirmwares, uploadFirmware, deleteFirmware,
  getOTATasks, createOTATask, deleteOTATask,
  pauseOTATask, resumeOTATask, cancelOTATask,
  completeOTATask, retryFailedDevices
} from '@/api/ota'
import { getDevices } from '@/api/device'
import { UPGRADE_STATUS_LABELS, UPGRADE_END_REASON_LABELS } from '@/utils/constants'

const firmwares = ref<any[]>([])
const tasks = ref<any[]>([])
const devices = ref<any[]>([])
const showCreateTask = ref(false)
const creatingTask = ref(false)
const taskForm = ref<{ firmwareId: number | null; deviceIds: string[] }>({
  firmwareId: null,
  deviceIds: []
})

const hasEndedTask = computed(() =>
  tasks.value.some(t => t.endReason)
)

function formatDT(dt: string) {
  if (!dt) return '-'
  return new Date(dt).toLocaleString('zh-CN')
}

function formatSize(bytes: number) {
  if (!bytes) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function getStatusType(status: string) {
  const map: Record<string, string> = {
    pending: 'info',
    running: '',
    paused: 'warning',
    completed: 'success',
    cancelled: 'danger'
  }
  return map[status] || 'info'
}

/** 进度条状态：控制颜色 */
function getProgressStatus(row: any) {
  if (row.status === 'running' || row.status === 'pending') return undefined // 蓝色（默认）
  if (row.status === 'paused') return 'warning'  // 橙色
  if (row.status === 'cancelled') return 'exception'  // 红色
  // completed 状态
  if (row.progress >= 100 && row.failCount === 0) return 'success'  // 绿色（全部成功）
  if (row.failCount > 0) return 'warning'  // 橙色（部分失败）
  return 'success'
}

function selectAllDevices() {
  taskForm.value.deviceIds = devices.value.map(d => d.id)
}

async function handleFileChange(file: any) {
  const formData = new FormData()
  formData.append('file', file.raw)
  formData.append('name', file.name)
  try {
    await uploadFirmware(formData)
    ElMessage.success('上传成功')
    loadFirmwares()
  } catch (e) {
    ElMessage.error('上传失败')
  }
}

async function handleDeleteFirmware(id: number) {
  try {
    await ElMessageBox.confirm('确定删除该固件吗？', '提示', { type: 'warning' })
    await deleteFirmware(id)
    ElMessage.success('固件已删除')
    loadFirmwares()
  } catch (e: any) {
    if (e !== 'cancel' && e?.message !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

async function openCreateTask() {
  showCreateTask.value = true
  taskForm.value = { firmwareId: null, deviceIds: [] }
  if (devices.value.length === 0) {
    try {
      const res: any = await getDevices()
      const raw = res.data || res
      // 兼容分页返回 {list:[...]} 和数组返回 [...]
      let allDevices = raw?.list || (Array.isArray(raw) ? raw : [])
      // 过滤掉气象站等非逆变器设备（只保留设备ID含 INV 或名称含逆变器的设备）
      devices.value = allDevices.filter((d: any) => {
        const devId = (d.deviceId || d.device_id || '').toUpperCase()
        const name = (d.name || '').toUpperCase()
        return devId.includes('INV') || name.includes('逆变器') || devId.includes('8661')
      })
    } catch {}
  }
}

async function handleCreateTask() {
  if (!taskForm.value.firmwareId) {
    ElMessage.warning('请选择固件')
    return
  }
  if (taskForm.value.deviceIds.length === 0) {
    ElMessage.warning('请选择至少一台设备')
    return
  }
  creatingTask.value = true
  try {
    await createOTATask({
      firmwareId: taskForm.value.firmwareId!,
      deviceIds: taskForm.value.deviceIds
    })
    ElMessage.success('任务创建成功')
    showCreateTask.value = false
    await loadTasks()
    startPolling()
  } catch (e: any) {
    ElMessage.error(e?.message || '创建失败')
  } finally {
    creatingTask.value = false
  }
}

async function pauseTask(id: number) {
  try {
    await pauseOTATask(id)
    ElMessage.success('已暂停')
    loadTasks()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

async function resumeTask(id: number) {
  try {
    await resumeOTATask(id)
    ElMessage.success('已继续')
    loadTasks()
    startPolling()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

async function cancelTask(id: number) {
  try {
    await ElMessageBox.confirm('确定取消该升级任务吗？', '提示', { type: 'warning' })
    await cancelOTATask(id)
    ElMessage.success('已取消')
    loadTasks()
  } catch (e: any) {
    if (e !== 'cancel' && e?.message !== 'cancel') {
      ElMessage.error(e?.message || '操作失败')
    }
  }
}

async function completeTask(id: number) {
  try {
    await ElMessageBox.confirm('手动结束任务将停止剩余设备的升级，进度停在当前位置。确定吗？', '手动结束', { type: 'warning' })
    await completeOTATask(id)
    ElMessage.success('任务已结束')
    loadTasks()
  } catch (e: any) {
    if (e !== 'cancel' && e?.message !== 'cancel') {
      ElMessage.error(e?.message || '操作失败')
    }
  }
}

async function retryTask(id: number) {
  try {
    await ElMessageBox.confirm('将重新升级所有失败的设备，确定吗？', '重试失败设备', { type: 'warning' })
    await retryFailedDevices(id)
    ElMessage.success('重试已启动')
    loadTasks()
    startPolling()
  } catch (e: any) {
    if (e !== 'cancel' && e?.message !== 'cancel') {
      ElMessage.error(e?.message || '操作失败')
    }
  }
}

async function handleDeleteTask(id: number) {
  try {
    await ElMessageBox.confirm('确定删除该升级任务记录吗？删除后不可恢复。', '删除任务', { type: 'warning' })
    await deleteOTATask(id)
    ElMessage.success('任务已删除')
    loadTasks()
  } catch (e: any) {
    if (e !== 'cancel' && e?.message !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

async function loadFirmwares() {
  try {
    const res: any = await getFirmwares()
    firmwares.value = res.data || []
  } catch (e) {
    firmwares.value = []
  }
}

async function loadTasks() {
  try {
    const res: any = await getOTATasks()
    tasks.value = res.data || []
  } catch (e) {
    tasks.value = []
  }
}

// 任务轮询:有运行中任务时每1.5s刷新,否则每5s
let pollTimer: any = null
function hasRunningTask() {
  return tasks.value.some((t: any) => t.status === 'pending' || t.status === 'running')
}
function startPolling() {
  stopPolling()
  const interval = hasRunningTask() ? 1500 : 5000
  pollTimer = setInterval(async () => {
    await loadTasks()
    if (!hasRunningTask() && pollTimer) {
      stopPolling()
      startPolling()
    }
  }, interval)
}
function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onMounted(() => {
  loadFirmwares()
  loadTasks().then(startPolling)
})
onUnmounted(stopPolling)
</script>

<style scoped lang="scss">
.ota-page {
  color: #e6ecf5;
}

.section {
  background: #182434;
  border: 1px solid #253650;
  border-radius: 12px;
  margin-bottom: 16px;
  overflow: hidden;

  .section-header {
    padding: 14px 18px;
    border-bottom: 1px solid #253650;
    font-size: 13px;
    font-weight: 600;
    background: #1d2b3e;
  }

  .section-body {
    padding: 16px 18px;
  }
}

.progress-detail {
  font-size: 11px;
  color: #566880;
  margin-top: 2px;
}

.end-reason {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
  &.all_success { background: rgba(29,201,129,0.15); color: #1dc981; }
  &.partial_fail { background: rgba(239,170,23,0.15); color: #efaa17; }
  &.all_failed { background: rgba(246,86,92,0.15); color: #f6565c; }
  &.manual_stop { background: rgba(77,163,255,0.15); color: #4da3ff; }
  &.cancelled { background: rgba(246,86,92,0.15); color: #f6565c; }
}

.device-select-area {
  border: 1px solid #253650;
  border-radius: 8px;
  padding: 12px;
  background: #111b2a;
  width: 100%;
  max-height: 240px;
  overflow-y: auto;

  .select-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 10px;

    .selected-count {
      margin-left: auto;
      font-size: 12px;
      color: #4da3ff;
    }
  }

  .device-checkboxes {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .device-status-tag {
    font-size: 10px;
    padding: 1px 6px;
    border-radius: 4px;
    margin-left: 6px;
    &.online { background: rgba(61,214,140,0.15); color: #3dd68c; }
    &.offline { background: rgba(246,86,92,0.15); color: #f6565c; }
  }
}

:deep(.el-table) {
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-header-bg-color: #1d2b3e;
  --el-table-border-color: #253650;
  --el-table-text-color: #e6ecf5;
  --el-table-header-text-color: #566880;
}

:deep(.el-dialog) {
  --el-dialog-bg-color: #182434;
  --el-dialog-title-font-size: 15px;
  border: 1px solid #253650;
  border-radius: 12px;
}

:deep(.el-checkbox__label) {
  color: #e6ecf5;
}
</style>
