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
        </el-table>
      </div>
    </div>

    <div class="section">
      <div class="section-header">升级任务</div>
      <div class="section-body">
        <el-button type="primary" size="small" @click="openCreateTask">创建升级任务</el-button>
        <el-table :data="tasks" size="small" stripe table-layout="auto" style="margin-top: 12px">
          <el-table-column prop="id" label="ID" />
          <el-table-column label="固件">
            <template #default="{ row }">{{ row.firmware?.name || '-' }}</template>
          </el-table-column>
          <el-table-column prop="status" label="状态">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)" size="small">{{ UPGRADE_STATUS_LABELS[row.status] || row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="进度">
            <template #default="{ row }">
              <el-progress :percentage="Math.round(row.progress)" :stroke-width="14" />
            </template>
          </el-table-column>
          <el-table-column label="成功/失败" align="center">
            <template #default="{ row }">{{ row.successCount }}/{{ row.failCount }}</template>
          </el-table-column>
          <el-table-column label="操作" align="center">
            <template #default="{ row }">
              <el-button v-if="row.status === 'pending' || row.status === 'paused'" link type="primary" size="small" @click="pauseTask(row.id)">暂停</el-button>
              <el-button v-if="row.status === 'running'" link type="primary" size="small" @click="resumeTask(row.id)">继续</el-button>
              <el-button v-if="row.status !== 'cancelled'" link type="danger" size="small" @click="cancelTask(row.id)">取消</el-button>
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
              <el-checkbox v-for="dev in devices" :key="dev.id" :label="dev.id" :value="dev.id">
                {{ dev.id }} <span class="device-status-tag" :class="dev.status">{{ dev.status === 'online' ? '在线' : '离线' }}</span>
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
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getFirmwares, uploadFirmware, getOTATasks, createOTATask, pauseOTATask, resumeOTATask, cancelOTATask } from '@/api/ota'
import { getDevices } from '@/api/device'
import { UPGRADE_STATUS_LABELS } from '@/utils/constants'

const firmwares = ref<any[]>([])
const tasks = ref<any[]>([])
const devices = ref<any[]>([])
const showCreateTask = ref(false)
const creatingTask = ref(false)
const taskForm = ref<{ firmwareId: number | null; deviceIds: string[] }>({
  firmwareId: null,
  deviceIds: []
})

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

async function openCreateTask() {
  showCreateTask.value = true
  taskForm.value = { firmwareId: null, deviceIds: [] }
  if (devices.value.length === 0) {
    try {
      const res: any = await getDevices()
      devices.value = res.data?.items || res.data || []
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
    startPolling()  // 立即启动快速轮询跟踪进度
  } catch (e: any) {
    ElMessage.error(e?.message || '创建失败')
  } finally {
    creatingTask.value = false
  }
}

async function pauseTask(id: number) {
  await pauseOTATask(id)
  ElMessage.success('已暂停')
  loadTasks()
}

async function resumeTask(id: number) {
  await resumeOTATask(id)
  ElMessage.success('已继续')
  loadTasks()
}

async function cancelTask(id: number) {
  await cancelOTATask(id)
  ElMessage.success('已取消')
  loadTasks()
}

async function loadFirmwares() {
  const res: any = await getFirmwares()
  firmwares.value = res.data || []
}

async function loadTasks() {
  const res: any = await getOTATasks()
  tasks.value = res.data || []
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
    // 状态变化后调整轮询频率
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
