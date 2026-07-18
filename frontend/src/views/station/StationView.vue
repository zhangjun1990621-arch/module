<template>
  <div class="station-page">
    <div class="section">
      <div class="section-header">
        台区管理
        <el-button type="primary" size="small" @click="showCreate = true">新建台区</el-button>
      </div>
      <div class="section-body">
        <el-table :data="stations" size="small" stripe table-layout="auto" v-loading="loading">
          <el-table-column prop="id" label="ID" />
          <el-table-column prop="name" label="台区名称" />
          <el-table-column prop="code" label="台区编号" />
          <el-table-column prop="region" label="所属地区" />
          <el-table-column prop="deviceCount" label="设备数" align="center" />
          <el-table-column label="操作" align="center">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="editStation(row)">编辑</el-button>
              <el-button link type="danger" size="small" @click="deleteStation(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <div class="section">
      <div class="section-header">分组管理</div>
      <div class="section-body">
        <el-empty description="分组管理功能开发中" />
      </div>
    </div>

    <el-dialog v-model="showCreate" :title="editingId ? '编辑台区' : '新建台区'" width="480px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="台区名称">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="台区编号">
          <el-input v-model="form.code" />
        </el-form-item>
        <el-form-item label="所属地区">
          <el-input v-model="form.region" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" @click="handleSave">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getStations, createStation, updateStation, deleteStation as apiDeleteStation } from '@/api/station'

const stations = ref<any[]>([])
const loading = ref(false)
const showCreate = ref(false)
const editingId = ref<number | null>(null)
const form = ref({ name: '', code: '', region: '' })

async function loadStations() {
  loading.value = true
  try {
    const res: any = await getStations()
    stations.value = res.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function editStation(station: any) {
  editingId.value = station.id
  form.value = { name: station.name, code: station.code, region: station.region }
  showCreate.value = true
}

async function handleSave() {
  try {
    if (editingId.value) {
      await updateStation(editingId.value, form.value)
    } else {
      await createStation(form.value)
    }
    ElMessage.success('保存成功')
    showCreate.value = false
    editingId.value = null
    form.value = { name: '', code: '', region: '' }
    loadStations()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  }
}

async function deleteStation(id: number) {
  await ElMessageBox.confirm('确认删除该台区？', '提示', { type: 'warning' })
  await apiDeleteStation(id)
  ElMessage.success('删除成功')
  loadStations()
}

onMounted(loadStations)
</script>

<style scoped lang="scss">
.station-page {
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
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .section-body {
    padding: 16px 18px;
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
  border: 1px solid #253650;
  border-radius: 12px;
}
</style>
