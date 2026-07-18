<template>
  <div class="al-history">
    <div class="al-card">
      <div class="al-card-header">
        <span class="al-card-title">历史数据</span>
        <div class="al-card-actions">
          <span style="font-size:12px;color:var(--text-tertiary)">共 {{ filtered.length }} 条</span>
          <el-button type="primary" size="small" @click="load">刷新</el-button>
        </div>
      </div>

      <div class="al-searchbar">
        <span class="al-search-label">电解槽</span>
        <el-select v-model="filterCell" placeholder="全部电解槽" clearable style="width: 150px" @change="page = 1">
          <el-option v-for="c in cellOptions" :key="c" :label="`${c}#`" :value="c" />
        </el-select>

        <el-input v-model="keyword" placeholder="搜索点位名称" clearable style="width: 200px" @input="page = 1" />

        <span style="margin-left:auto;font-size:12px;color:var(--text-tertiary)">
          对应原 T_DeviceData（按天分表）
        </span>
      </div>

      <div class="al-card-body" style="padding: 0">
        <el-table :data="paged" size="small" stripe table-layout="auto">
          <el-table-column prop="time" label="时间" width="165" align="center" />
          <el-table-column prop="cellName" label="槽号" align="center" />
          <el-table-column prop="positionName" label="位置" align="center" />
          <el-table-column prop="pointName" label="点位" align="center" />
          <el-table-column label="温度" align="center">
            <template #default="{ row }"><span :style="tempMeta(row.temp)">{{ row.temp }}<span style="font-size:12px">℃</span></span></template>
          </el-table-column>
          <el-table-column label="电压" align="center">
            <template #default="{ row }">{{ row.volt }}<span style="font-size:12px">V</span></template>
          </el-table-column>
          <el-table-column label="电流" align="center">
            <template #default="{ row }">{{ row.current }}<span style="font-size:12px">A</span></template>
          </el-table-column>
        </el-table>
        <div style="display:flex;justify-content:flex-end;padding:12px 16px">
          <el-pagination v-model:current-page="page" :page-size="pageSize" :total="filtered.length" layout="prev, pager, next, total" small background />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { aluminumApi } from './mockData'
import { tempMeta } from './format'
import type { HistoryRow } from './types'
import './aluminum.scss'

const history = ref<HistoryRow[]>([])
const filterCell = ref<number | ''>('')
const keyword = ref<string>('')
const page = ref(1)
const pageSize = 15

const cellOptions = computed(() => Array.from(new Set(history.value.map((h) => h.cellId))).sort((a, b) => a - b))

const filtered = computed(() =>
  history.value.filter((h) => {
    if (filterCell.value !== '' && h.cellId !== filterCell.value) return false
    if (keyword.value && !h.pointName.includes(keyword.value)) return false
    return true
  })
)
const paged = computed(() => filtered.value.slice((page.value - 1) * pageSize, page.value * pageSize))

async function load() {
  history.value = await aluminumApi.getHistory()
}

onMounted(load)
</script>
