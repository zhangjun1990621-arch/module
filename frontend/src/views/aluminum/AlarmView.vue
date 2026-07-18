<template>
  <div class="al-alarms">
    <div class="al-card">
      <div class="al-card-header">
        <span class="al-card-title">报警记录</span>
        <div class="al-card-actions">
          <el-button size="small" @click="handleExport">导出CSV</el-button>
        </div>
      </div>

      <div class="al-searchbar">
        <span class="al-search-label">等级</span>
        <el-select v-model="filterLevel" style="width: 120px" @change="page = 1">
          <el-option label="全部" :value="''" />
          <el-option label="一级" :value="1" />
          <el-option label="二级" :value="2" />
          <el-option label="三级" :value="3" />
        </el-select>

        <span class="al-search-label">类型</span>
        <el-select v-model="filterType" style="width: 150px" @change="page = 1">
          <el-option label="全部" value="" />
          <el-option label="槽温超限" value="槽温超限" />
          <el-option label="升温趋势" value="升温趋势" />
          <el-option label="破损高危" value="破损高危" />
          <el-option label="钢棒切削" value="钢棒切削" />
        </el-select>

        <span class="al-search-label">状态</span>
        <el-select v-model="filterStatus" style="width: 130px" @change="page = 1">
          <el-option label="全部" value="" />
          <el-option label="未处理" value="未处理" />
          <el-option label="处理中" value="处理中" />
          <el-option label="已恢复" value="已恢复" />
        </el-select>

        <el-input v-model="keyword" placeholder="搜索槽号/点位/详情" clearable style="width: 220px" @input="page = 1" />

        <span style="margin-left: auto; font-size: 12px; color: var(--text-tertiary)">
          共 {{ filtered.length }} 条
        </span>
      </div>

      <div class="al-card-body" style="padding: 0">
        <el-table :data="paged" size="small" stripe table-layout="auto">
          <el-table-column prop="time" label="时间" width="165" align="center" />
          <el-table-column prop="cellName" label="槽号" align="center" />
          <el-table-column prop="pointName" label="点位名称" align="center" />
          <el-table-column label="等级" align="center">
            <template #default="{ row }">
              <span class="al-tag" :class="levelMeta(row.level).cls">{{ levelMeta(row.level).label }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="type" label="类型" align="center" />
          <el-table-column prop="detail" label="详情" align="center" />
          <el-table-column label="状态" align="center">
            <template #default="{ row }">
              <span class="al-tag" :class="alarmStatusMeta(row.status).cls">{{ alarmStatusMeta(row.status).label }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" fixed="right" align="center">
            <template #default="{ row }">
              <el-button text type="primary" size="small" @click="goCell(row.cellId)">查看槽</el-button>
            </template>
          </el-table-column>
        </el-table>

        <div style="display: flex; justify-content: flex-end; padding: 12px 16px">
          <el-pagination
            v-model:current-page="page"
            v-model:page-size="pageSize"
            :total="filtered.length"
            :page-sizes="[20, 50, 100]"
            layout="total, sizes, prev, pager, next"
            background
            @size-change="page = 1"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { aluminumApi } from './mockData'
import { levelMeta, alarmStatusMeta } from './format'
import { exportToCSV } from './exportUtil.ts'
import type { AlarmRecord } from './types'
import './aluminum.scss'

const router = useRouter()
const alarms = ref<AlarmRecord[]>([])
const filterLevel = ref<number | ''>('')
const filterType = ref<string>('')
const filterStatus = ref<string>('')
const keyword = ref<string>('')
const page = ref(1)
const pageSize = ref(20)

const filtered = computed(() =>
  alarms.value.filter((a) => {
    if (filterLevel.value !== '' && a.level !== filterLevel.value) return false
    if (filterType.value && a.type !== filterType.value) return false
    if (filterStatus.value && a.status !== filterStatus.value) return false
    if (keyword.value) {
      const kw = keyword.value.trim().toLowerCase()
      const hay = `${a.cellName} ${a.pointName} ${a.detail} ${a.type}`.toLowerCase()
      if (!hay.includes(kw)) return false
    }
    return true
  })
)

const paged = computed(() => filtered.value.slice((page.value - 1) * pageSize.value, page.value * pageSize.value))

function goCell(cellId: number) {
  router.push(`/al/points?cell=${cellId}`)
}

/** 导出当前筛选结果为 CSV */
function handleExport() {
  const headers = [
    { label: '时间', prop: 'time' },
    { label: '槽号', prop: 'cellName' },
    { label: '点位名称', prop: 'pointName' },
    { label: '等级', prop: 'levelLabel' },
    { label: '类型', prop: 'type' },
    { label: '详情', prop: 'detail' },
    { label: '状态', prop: 'statusLabel' }
  ]
  const data = filtered.value.map((a) => ({
    time: a.time,
    cellName: a.cellName,
    pointName: a.pointName,
    levelLabel: levelMeta(a.level).label,
    type: a.type,
    detail: a.detail,
    statusLabel: alarmStatusMeta(a.status).label
  }))
  exportToCSV('报警记录', headers, data)
}

async function load() {
  alarms.value = await aluminumApi.getAlarms()
}

onMounted(load)
</script>
