<template>
  <div class="al-points">
    <div class="al-card">
      <div class="al-card-header">
        <span class="al-card-title">点位实时数据</span>
        <div class="al-card-actions">
          <el-button size="small" @click="handleExport">导出CSV</el-button>
          <el-button type="primary" size="small" @click="load">刷新</el-button>
        </div>
      </div>

      <!-- 四级级联筛选：厂区 → 工区 → 设备主机 → 槽号 + 位置 + 异常标记 -->
      <div class="al-searchbar">
        <span class="al-search-label">厂区</span>
        <el-select v-model="filterFactory" placeholder="全部厂区" clearable style="width: 120px" @change="onFactoryChange">
          <el-option v-for="f in factoryOptions" :key="f" :label="f" :value="f" />
        </el-select>

        <span class="al-search-label">工区</span>
        <el-select v-model="filterWorkzone" placeholder="全部工区" clearable style="width: 120px" @change="onWorkzoneChange">
          <el-option v-for="w in workzoneOptions" :key="w" :label="w" :value="w" />
        </el-select>

        <span class="al-search-label">主机</span>
        <el-select v-model="filterHost" placeholder="全部主机" clearable style="width: 120px" @change="onHostChange">
          <el-option v-for="h in hostOptions" :key="h" :label="h" :value="h" />
        </el-select>

        <span class="al-search-label">槽号</span>
        <el-select v-model="filterCell" placeholder="全部槽号" clearable filterable style="width: 130px" @change="page = 1">
          <el-option v-for="c in cellOptions" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>

        <span class="al-search-label">位置</span>
        <el-select v-model="filterPos" placeholder="全部位置" clearable style="width: 120px" @change="page = 1">
          <el-option v-for="p in positions" :key="p.code" :label="p.name" :value="p.code" />
        </el-select>

        <span class="al-search-label">异常</span>
        <el-select v-model="filterFlag" placeholder="全部" clearable style="width: 110px" @change="page = 1">
          <el-option label="破损" value="damaged" />
          <el-option label="切削" value="cut" />
          <el-option label="报警" value="alarm" />
        </el-select>

        <span style="margin-left: auto; font-size: 12px; color: var(--text-tertiary)">
          共 {{ filtered.length }} 个点位
        </span>
      </div>

      <div class="al-card-body" style="padding: 0">
        <el-table :data="paged" size="small" stripe>
          <el-table-column prop="id" label="点位ID" width="90" />
          <el-table-column prop="cellName" label="槽号" width="60" align="center" />
          <el-table-column label="厂区" width="60" align="center">
            <template #default="{ row }">{{ cellMap.get(row.cellId)?.factory || '—' }}</template>
          </el-table-column>
          <el-table-column label="工区" width="60" align="center">
            <template #default="{ row }">{{ cellMap.get(row.cellId)?.workzone || '—' }}</template>
          </el-table-column>
          <el-table-column label="主机" width="55" align="center">
            <template #default="{ row }">{{ cellMap.get(row.cellId)?.host || '—' }}</template>
          </el-table-column>
          <el-table-column prop="positionName" label="位置" width="70" align="center" />
          <el-table-column prop="name" label="点位名称" width="85" />
          <el-table-column label="温度" width="70" align="right">
            <template #default="{ row }">
              <span :style="tempMeta(row.temp)">{{ row.temp }}<span style="font-size: 12px">℃</span></span>
            </template>
          </el-table-column>
          <el-table-column label="电压V" width="60" align="right">
            <template #default="{ row }">{{ row.volt }}<span style="font-size: 12px">V</span></template>
          </el-table-column>
          <el-table-column label="电流A" width="65" align="right">
            <template #default="{ row }">{{ row.current }}<span style="font-size: 12px">A</span></template>
          </el-table-column>
          <el-table-column label="升温趋势" width="65" align="center">
            <template #default="{ row }">
              <span class="al-tag" :class="levelMeta(row.warmingLevel).cls">{{ levelMeta(row.warmingLevel).label }}</span>
            </template>
          </el-table-column>
          <el-table-column label="报警等级" width="65" align="center">
            <template #default="{ row }">
              <span class="al-tag" :class="levelMeta(row.errorLevel).cls">{{ levelMeta(row.errorLevel).label }}</span>
            </template>
          </el-table-column>
          <el-table-column label="破损/切削" width="70" align="center">
            <template #default="{ row }">
              <span v-if="row.damaged" class="al-tag red">破损</span>
              <span v-if="row.cut" class="al-tag orange">切削</span>
              <span v-if="!row.damaged && !row.cut" style="color: #8d9db8">—</span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="65" align="center">
            <template #default="{ row }">
              <span class="al-dot" :class="row.status === 'online' ? 'online' : row.status === 'offline' ? 'offline' : 'alarm'" />
              <span class="al-tag" :class="pointStatusMeta(row.status).cls">{{ pointStatusMeta(row.status).label }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="saveTime" label="采集时间" width="130" align="center" />
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
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { aluminumApi, POSITIONS } from './mockData'
import { pointStatusMeta, levelMeta, tempMeta } from './format'
import { exportToCSV } from './exportUtil.ts'
import type { CellPoint, Cell } from './types'
import './aluminum.scss'

const route = useRoute()
const points = ref<CellPoint[]>([])
const cells = ref<Cell[]>([])
const positions = POSITIONS

/* 四级级联筛选 */
const filterFactory = ref<string>('')
const filterWorkzone = ref<string>('')
const filterHost = ref<string>('')
const filterCell = ref<number | ''>('')
const filterPos = ref<string>('')
const filterFlag = ref<string>('')
const page = ref(1)
const pageSize = ref(20)

/* cellId → Cell 映射，用于点位表显示厂区/工区/主机 */
const cellMap = computed(() => {
  const m = new Map<number, Cell>()
  for (const c of cells.value) m.set(c.id, c)
  return m
})

/* 级联选项计算 */
const factoryOptions = computed(() => Array.from(new Set(cells.value.map((c) => c.factory))))

const workzoneOptions = computed(() => {
  const list = cells.value.filter((c) => !filterFactory.value || c.factory === filterFactory.value)
  return Array.from(new Set(list.map((c) => c.workzone)))
})

const hostOptions = computed(() => {
  const list = cells.value.filter(
    (c) => (!filterFactory.value || c.factory === filterFactory.value) &&
           (!filterWorkzone.value || c.workzone === filterWorkzone.value)
  )
  return Array.from(new Set(list.map((c) => c.host)))
})

const cellOptions = computed(() => {
  return cells.value.filter(
    (c) => (!filterFactory.value || c.factory === filterFactory.value) &&
           (!filterWorkzone.value || c.workzone === filterWorkzone.value) &&
           (!filterHost.value || c.host === filterHost.value)
  )
})

/* 级联重置 */
function onFactoryChange() {
  filterWorkzone.value = ''
  filterHost.value = ''
  filterCell.value = ''
  page.value = 1
}
function onWorkzoneChange() {
  filterHost.value = ''
  filterCell.value = ''
  page.value = 1
}
function onHostChange() {
  filterCell.value = ''
  page.value = 1
}

/* 最终筛选 */
const filtered = computed(() => {
  // 先确定满足级联条件的 cellId 集合
  const allowedCellIds = new Set(
    cellOptions.value
      .filter((c) => !filterCell.value || c.id === filterCell.value)
      .map((c) => c.id)
  )
  return points.value.filter((p) => {
    if (!allowedCellIds.has(p.cellId)) return false
    if (filterPos.value && p.position !== filterPos.value) return false
    if (filterFlag.value === 'damaged' && !p.damaged) return false
    if (filterFlag.value === 'cut' && !p.cut) return false
    if (filterFlag.value === 'alarm' && p.errorLevel < 1) return false
    return true
  })
})

const paged = computed(() => filtered.value.slice((page.value - 1) * pageSize.value, page.value * pageSize.value))

/** 导出当前筛选结果为 CSV */
function handleExport() {
  const headers = [
    { label: '点位ID', prop: 'id' },
    { label: '槽号', prop: 'cellName' },
    { label: '厂区', prop: 'factory' },
    { label: '工区', prop: 'workzone' },
    { label: '设备主机', prop: 'host' },
    { label: '位置', prop: 'positionName' },
    { label: '点位名称', prop: 'name' },
    { label: '温度(℃)', prop: 'temp' },
    { label: '电压(V)', prop: 'volt' },
    { label: '电流(A)', prop: 'current' },
    { label: '升温趋势', prop: 'warmingLabel' },
    { label: '报警等级', prop: 'errorLabel' },
    { label: '破损', prop: 'damagedText' },
    { label: '切削', prop: 'cutText' },
    { label: '状态', prop: 'statusLabel' },
    { label: '采集时间', prop: 'saveTime' }
  ]
  const data = filtered.value.map((p) => {
    const cell = cellMap.value.get(p.cellId)
    return {
      id: p.id,
      cellName: p.cellName,
      factory: cell?.factory || '—',
      workzone: cell?.workzone || '—',
      host: cell?.host || '—',
      positionName: p.positionName,
      name: p.name,
      temp: p.temp,
      volt: p.volt,
      current: p.current,
      warmingLabel: levelMeta(p.warmingLevel).label,
      errorLabel: levelMeta(p.errorLevel).label,
      damagedText: p.damaged ? '破损' : '—',
      cutText: p.cut ? '切削' : '—',
      statusLabel: pointStatusMeta(p.status).label,
      saveTime: p.saveTime
    }
  })
  exportToCSV('点位实时数据', headers, data)
}

async function load() {
  const [ps, cs] = await Promise.all([aluminumApi.getPoints(), aluminumApi.getCells()])
  points.value = ps
  cells.value = cs
  const q = route.query.cell
  if (q) filterCell.value = Number(q)
}

// 从概览/电解槽/报警页跳转携带 ?cell= 变化时同步筛选
watch(
  () => route.query.cell,
  (q) => {
    filterCell.value = q ? Number(q) : ''
    page.value = 1
  }
)

onMounted(load)
</script>

<style scoped>
:deep(.el-table__body),
:deep(.el-table__header) {
  width: 100% !important;
}
</style>
