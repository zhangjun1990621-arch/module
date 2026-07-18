<template>
  <div class="al-cells">
    <div class="al-card">
      <div class="al-card-header">
        <span class="al-card-title">电解槽监测</span>
      </div>

      <!-- 四级级联筛选：厂区 → 工区 → 主机 → 槽 -->
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
        <el-select v-model="filterCell" placeholder="请选择槽" filterable style="width: 130px" @change="onCellChange">
          <el-option v-for="c in cellOptions" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>

        <div style="margin-left: auto">
          <el-button size="small" @click="loadMatrix">刷新</el-button>
        </div>
      </div>

      <!-- 单槽温度矩阵 4行×28列，每行上方写死标签 -->
      <div class="al-card-body" style="padding: 0; overflow-x: auto">
        <div v-if="!filterCell" class="empty-tip">请选择槽号查看温度矩阵</div>
        <table v-else class="temp-matrix-table">
          <tbody>
            <template v-for="(rowArr, ri) in matrix" :key="ri">
              <!-- 标签行 -->
              <tr class="label-row">
                <td
                  v-for="col in 28"
                  :key="'label-' + ri + '-' + col"
                  class="label-cell"
                >{{ getLabel(ri, col) }}</td>
              </tr>
              <!-- 数据行 -->
              <tr>
                <td
                  v-for="cell in rowArr"
                  :key="cell.row + '-' + cell.col"
                  class="temp-cell"
                  :class="getCellClass(cell)"
                  :title="`${getLabel(ri, cell.col)}: ${cell.temp > 0 ? cell.temp + '℃' : '离线'}`"
                >
                  {{ cell.temp > 0 ? cell.temp : '—' }}
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog
      v-model="detailVisible"
      :title="detailTitle"
      width="80%"
      top="6vh"
      class="al-detail-dialog"
      destroy-on-close
    >
      <el-tabs v-model="activeTab" type="border-card" tab-position="left">
        <el-tab-pane label="历史曲线" name="chart">
          <div class="al-tab-tip">近 12 小时槽温变化趋势</div>
          <AlChart :option="historyChartOption" height="360px" />
        </el-tab-pane>
        <el-tab-pane :label="`实时报警 (${currentAlarms.length})`" name="alarms">
          <el-table :data="currentAlarms" size="small" stripe :max-height="420">
            <el-table-column prop="time" label="时间" width="160" />
            <el-table-column prop="id" label="报警ID" width="110" />
            <el-table-column prop="pointName" label="点位" width="120" />
            <el-table-column label="等级" width="80" align="center">
              <template #default="{ row }">
                <span class="al-tag" :class="levelMeta(row.level).cls">{{ levelMeta(row.level).label }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="type" label="类型" width="100" />
            <el-table-column prop="detail" label="详情" min-width="120" />
            <el-table-column label="处理状态" width="100" align="center">
              <template #default="{ row }">
                <span class="al-tag" :class="alarmStatusMeta(row.status).cls">{{ row.status }}</span>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!currentAlarms.length" description="该电解槽暂无活跃报警" />
        </el-tab-pane>
        <el-tab-pane :label="`点位信息 (${currentPoints.length})`" name="points">
          <el-table :data="currentPoints" size="small" stripe :max-height="420">
            <el-table-column prop="positionName" label="位置" width="90" />
            <el-table-column prop="name" label="点位名称" width="120" />
            <el-table-column label="温度" width="100" align="right">
              <template #default="{ row }">
                <span :style="tempMeta(row.temp)">{{ row.temp }}<span class="al-unit">℃</span></span>
              </template>
            </el-table-column>
            <el-table-column label="电压" width="90" align="right">
              <template #default="{ row }">{{ row.volt }}<span class="al-unit">V</span></template>
            </el-table-column>
            <el-table-column label="电流" width="90" align="right">
              <template #default="{ row }">{{ row.current }}<span class="al-unit">A</span></template>
            </el-table-column>
            <el-table-column label="状态" width="80" align="center">
              <template #default="{ row }">
                <span class="al-tag" :class="pointStatusMeta(row.status).cls">{{ pointStatusMeta(row.status).label }}</span>
              </template>
            </el-table-column>
            <el-table-column label="升温趋势" width="90" align="center">
              <template #default="{ row }">
                <span class="al-tag" :class="levelMeta(row.warmingLevel).cls">{{ levelMeta(row.warmingLevel).label }}</span>
              </template>
            </el-table-column>
            <el-table-column label="三级报警" width="90" align="center">
              <template #default="{ row }">
                <span class="al-tag" :class="levelMeta(row.errorLevel).cls">{{ levelMeta(row.errorLevel).label }}</span>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!currentPoints.length" description="该电解槽暂无点位数据" />
        </el-tab-pane>
      </el-tabs>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import type * as echarts from 'echarts'
import AlChart from './AlChart.vue'
import { aluminumApi } from './mockData'
import { cellStatusMeta, pointStatusMeta, levelMeta, alarmStatusMeta, tempMeta } from './format'
import type { Cell, CellPoint, AlarmRecord, TempMatrixCell } from './types'
import './aluminum.scss'

const route = useRoute()

/* ---------------- 矩阵数据 ---------------- */
const matrix = ref<TempMatrixCell[][]>([])

/* 4行定义：A面1层、A面2层、B面1层、B面2层 */
const rowDefs = [
  { side: 'A', layer: 1 },
  { side: 'A', layer: 2 },
  { side: 'B', layer: 1 },
  { side: 'B', layer: 2 }
]

/* 生成标签：A1-1, A2-1, ..., A28-1 / A1-2, A2-2, ..., A28-2 / B1-1... / B1-2... */
function getLabel(ri: number, col: number): string {
  return `${rowDefs[ri].side}${col}-${rowDefs[ri].layer}`
}

/* ---------------- 列表数据 ---------------- */
const cells = ref<Cell[]>([])
const allPoints = ref<CellPoint[]>([])
const allAlarms = ref<AlarmRecord[]>([])

/* 四级级联筛选 */
const filterFactory = ref<string>('')
const filterWorkzone = ref<string>('')
const filterHost = ref<string>('')
const filterCell = ref<number | ''>('')

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
}
function onWorkzoneChange() {
  filterHost.value = ''
  filterCell.value = ''
}
function onHostChange() {
  filterCell.value = ''
}
function onCellChange() {
  if (filterCell.value) loadMatrix()
}

/* ---------------- 矩阵加载 ---------------- */
async function loadMatrix() {
  if (!filterCell.value) {
    matrix.value = []
    return
  }
  try {
    matrix.value = await aluminumApi.getCellTempMatrix(filterCell.value)
  } catch (e) {
    console.error('加载温度矩阵失败', e)
  }
}

function getCellClass(cell: TempMatrixCell): string {
  if (cell.temp === 0) return 'bg-dark'
  if (cell.errorLevel === 3) return 'bg-red'
  if (cell.errorLevel === 2) return 'bg-orange'
  if (cell.errorLevel === 1) return 'bg-yellow'
  return 'bg-blue'
}

/* ---------------- 5秒静默自动刷新 ---------------- */
let refreshTimer: ReturnType<typeof setInterval> | null = null

function startAutoRefresh() {
  stopAutoRefresh()
  refreshTimer = setInterval(() => {
    if (filterCell.value) loadMatrix()
  }, 5000)
}

function stopAutoRefresh() {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

/* ---------------- 详情弹窗 ---------------- */
const detailVisible = ref(false)
const currentCell = ref<Cell | null>(null)
const activeTab = ref<'chart' | 'alarms' | 'points'>('chart')

const detailTitle = computed(() => (currentCell.value ? `${currentCell.value.name} 电解槽详情` : '电解槽详情'))

const currentPoints = computed(() =>
  currentCell.value ? allPoints.value.filter((p) => p.cellId === currentCell.value!.id) : []
)
const currentAlarms = computed(() =>
  currentCell.value ? allAlarms.value.filter((a) => a.cellId === currentCell.value!.id) : []
)

function openDetail(row: Cell) {
  currentCell.value = row
  activeTab.value = 'chart'
  detailVisible.value = true
}

/* ---------------- 演示曲线 ---------------- */
function makeRng(seed: number) {
  let s = seed >>> 0
  return () => {
    s = (s * 1664525 + 1013904223) >>> 0
    return s / 4294967296
  }
}
const pad2 = (n: number) => String(n).padStart(2, '0')

function genCellHistory(cell: Cell): { time: string; temp: number }[] {
  const rng = makeRng(cell.id * 13 + 7)
  const drift = cell.status === 'alarm' ? 12 : cell.status === 'warn' ? 7 : cell.status === 'offline' ? 0 : 2
  const base = cell.avgTemp > 0 ? cell.avgTemp : 940
  const now = new Date()
  const list: { time: string; temp: number }[] = []
  for (let i = 12; i >= 0; i--) {
    const t = new Date(now.getTime() - i * 3600 * 1000)
    const progress = (12 - i) / 12
    const noise = (rng() - 0.5) * 5
    list.push({ time: `${pad2(t.getHours())}:00`, temp: +(base + drift * progress + noise).toFixed(1) })
  }
  return list
}

const historyChartOption = computed<echarts.EChartsOption>(() => {
  const cell = currentCell.value
  if (!cell) return {}
  const series = genCellHistory(cell)
  return {
    backgroundColor: '#1e2a3e',
    title: {
      text: `${cell.name} 近 12 小时温度趋势`,
      subtext: `平均槽温 ${cell.avgTemp}℃ · 最高 ${cell.maxTemp}℃ · 状态：${cellStatusMeta(cell.status).label}`,
      left: 'center',
      textStyle: { color: '#e6ecf5', fontSize: 14 },
      subtextStyle: { color: '#8d9db8', fontSize: 12 }
    },
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(24,36,52,0.95)',
      borderColor: '#253650',
      textStyle: { color: '#e6ecf5', fontSize: 12 },
      valueFormatter: (v: any) => `${v}℃`
    },
    grid: { left: 56, right: 24, top: 72, bottom: 40 },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: series.map((d) => d.time),
      axisLine: { lineStyle: { color: '#2c3e55' } },
      axisLabel: { color: '#8d9db8', fontSize: 11 },
      axisTick: { show: false }
    },
    yAxis: {
      type: 'value',
      name: '温度(℃)',
      nameTextStyle: { color: '#8d9db8', fontSize: 11 },
      scale: true,
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: { lineStyle: { color: 'rgba(255,255,255,0.06)' } },
      axisLabel: { color: '#8d9db8', fontSize: 11, formatter: '{value}℃' }
    },
    series: [
      {
        name: '槽温',
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: 6,
        data: series.map((d) => d.temp),
        lineStyle: { color: '#1e9bf3', width: 2 },
        itemStyle: { color: '#1e9bf3' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(30,155,243,0.35)' },
              { offset: 1, color: 'rgba(30,155,243,0.02)' }
            ]
          }
        },
        markLine: {
          symbol: 'none',
          lineStyle: { type: 'dashed' },
          data: [
            { yAxis: 965, lineStyle: { color: '#f0a030' }, label: { formatter: '预警 965℃', color: '#f0a030', position: 'insideEndTop' } },
            { yAxis: 975, lineStyle: { color: '#f6565c' }, label: { formatter: '报警 975℃', color: '#f6565c', position: 'insideEndTop' } }
          ]
        }
      }
    ]
  }
})

/* ---------------- 数据加载 ---------------- */
async function load() {
  const [cs, ps, als] = await Promise.all([
    aluminumApi.getCells(),
    aluminumApi.getPoints(),
    aluminumApi.getAlarms()
  ])
  cells.value = cs
  allPoints.value = ps
  allAlarms.value = als
  const q = route.query.cell
  if (q) {
    const id = Number(q)
    const cell = cs.find((c) => c.id === id)
    if (cell) {
      filterCell.value = id
      loadMatrix()
    }
  }
}

onMounted(() => {
  load()
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped lang="scss">
.al-unit {
  font-size: 12px;
  opacity: 0.7;
  margin-left: 1px;
}
.al-tab-tip {
  font-size: 12px;
  color: #8d9db8;
  margin-bottom: 8px;
}
.empty-tip {
  text-align: center;
  padding: 60px 0;
  color: #8d9db8;
  font-size: 14px;
}

/* ---- 4×28 温度矩阵表 ---- */
.temp-matrix-table {
  border-collapse: collapse;
  width: 100%;
  font-size: 12px;
  font-family: 'Consolas', 'Monaco', 'Microsoft YaHei', monospace;

  .label-cell {
    text-align: center;
    padding: 4px 2px;
    font-size: 11px;
    font-weight: 600;
    color: #8d9db8;
    background: #131c2a;
    border: 1px solid #1a2a3e;
    min-width: 44px;
    white-space: nowrap;
  }

  .temp-cell {
    text-align: center;
    padding: 7px 2px;
    border: 1px solid #0d1420;
    color: #fff;
    font-weight: 600;
    font-size: 13px;
    min-width: 44px;
    height: 30px;
    transition: filter 0.15s;

    &:hover {
      filter: brightness(1.3);
      cursor: default;
    }

    &.bg-blue { background: rgba(4, 138, 235, 0.85); }
    &.bg-yellow { background: #e6c34a; color: #3a3a3a; }
    &.bg-orange { background: #f0a030; }
    &.bg-red { background: #f6565c; }
    &.bg-dark { background: #1a2333; color: #566880; }
  }

  /* 标签行和数据行之间无间距 */
  .label-row + tr td {
    border-top: none;
  }
}
</style>
