<template>
  <div class="al-overview">
    <div class="al-stat-grid">
      <div class="al-stat-card blue">
        <div class="al-stat-icon">🏭</div>
        <div>
          <div class="al-stat-label">电解槽总数</div>
          <div class="al-stat-value">{{ stat?.cellCount || 0 }}</div>
          <div class="al-stat-sub">在线 <span style="color:#3dd68c">{{ stat?.onlineCount || 0 }}</span> / 离线 <span style="color:#8d9db8">{{ stat?.offlineCount || 0 }}</span></div>
        </div>
      </div>
      <div class="al-stat-card red">
        <div class="al-stat-icon">🔔</div>
        <div>
          <div class="al-stat-label">活跃报警</div>
          <div class="al-stat-value">{{ stat?.alarmCount || 0 }}</div>
          <div class="al-stat-sub">一级 {{ stat?.level1 }} · 二级 {{ stat?.level2 }} · 三级 {{ stat?.level3 }}</div>
        </div>
      </div>
      <div class="al-stat-card orange">
        <div class="al-stat-icon">🌡️</div>
        <div>
          <div class="al-stat-label">平均槽温</div>
          <div class="al-stat-value">{{ stat?.avgTemp || 0 }}<span style="font-size:14px">℃</span></div>
          <div class="al-stat-sub">全部在线电解槽均值</div>
        </div>
      </div>
      <div class="al-stat-card green">
        <div class="al-stat-icon">🛡️</div>
        <div>
          <div class="al-stat-label">破损 / 切削</div>
          <div class="al-stat-value">{{ stat?.damagedCount || 0 }}<span style="font-size:14px"> / </span>{{ stat?.cutCount || 0 }}</div>
          <div class="al-stat-sub">破损高危 / 钢棒切削 点位</div>
        </div>
      </div>
    </div>

    <div class="al-row2">
      <div class="al-panel">
        <div class="al-panel-header">电解槽状态分布（按厂区）</div>
        <div class="al-panel-body">
          <div v-for="item in factoryStat" :key="item.factory" class="al-status-row">
            <span class="al-status-label">{{ item.factory }}</span>
            <el-progress
              :percentage="item.total > 0 ? Math.round((item.online / item.total) * 100) : 0"
              :stroke-width="18"
              :color="item.online / item.total >= 0.8 ? '#3dd68c' : item.online / item.total >= 0.5 ? '#f0a030' : '#f6565c'"
            />
            <span class="al-status-count" :style="{ color: item.online / item.total >= 0.8 ? '#3dd68c' : item.online / item.total >= 0.5 ? '#f0a030' : '#f6565c' }">
              {{ item.online }}/{{ item.total }}
            </span>
          </div>
          <el-empty v-if="!factoryStat.length" description="暂无数据" />
        </div>
      </div>

      <div class="al-panel">
        <div class="al-panel-header">设备主机状态</div>
        <div class="al-panel-body" style="padding:0">
          <el-table :data="hostStat" size="small" stripe>
            <el-table-column prop="host" label="主机" width="90" />
            <el-table-column prop="factory" label="厂区" width="80" />
            <el-table-column prop="workzone" label="工区" width="80" />
            <el-table-column label="在线率" width="120">
              <template #default="{ row }">
                <el-progress
                  :percentage="row.total > 0 ? Math.round((row.online / row.total) * 100) : 0"
                  :stroke-width="14"
                  :color="row.online / row.total >= 0.8 ? '#3dd68c' : row.online / row.total >= 0.5 ? '#f0a030' : '#f6565c'"
                />
              </template>
            </el-table-column>
            <el-table-column label="在线/总数" width="90" align="center">
              <template #default="{ row }">
                <span :style="{ color: row.online / row.total >= 0.8 ? '#3dd68c' : row.online / row.total >= 0.5 ? '#f0a030' : '#f6565c', fontWeight: 600 }">
                  {{ row.online }}/{{ row.total }}
                </span>
              </template>
            </el-table-column>
            <el-table-column label="平均槽温" width="100" align="right">
              <template #default="{ row }">
                <span :style="tempMeta(row.avgTemp)">{{ row.avgTemp }}℃</span>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!hostStat.length" description="暂无数据" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { aluminumApi } from './mockData'
import { tempMeta } from './format'
import type { OverviewStat, Cell } from './types'
import './aluminum.scss'

const stat = ref<OverviewStat | null>(null)
const cells = ref<Cell[]>([])

const factoryStat = computed(() => {
  const map = new Map<string, { factory: string; total: number; online: number }>()
  for (const c of cells.value) {
    const f = map.get(c.factory) || { factory: c.factory, total: 0, online: 0 }
    f.total++
    if (c.online) f.online++
    map.set(c.factory, f)
  }
  return Array.from(map.values())
})

/* 设备主机状态：按 host 分组统计在线率与平均槽温 */
const hostStat = computed(() => {
  const map = new Map<string, { host: string; factory: string; workzone: string; total: number; online: number; temps: number[] }>()
  for (const c of cells.value) {
    const h = map.get(c.host) || { host: c.host, factory: c.factory, workzone: c.workzone, total: 0, online: 0, temps: [] }
    h.total++
    if (c.online) {
      h.online++
      if (c.avgTemp > 0) h.temps.push(c.avgTemp)
    }
    map.set(c.host, h)
  }
  return Array.from(map.values()).map((h) => ({
    host: h.host,
    factory: h.factory,
    workzone: h.workzone,
    total: h.total,
    online: h.online,
    avgTemp: h.temps.length ? +(h.temps.reduce((a, b) => a + b, 0) / h.temps.length).toFixed(1) : 0
  }))
})

onMounted(async () => {
  try {
    const [s, c] = await Promise.all([aluminumApi.getOverview(), aluminumApi.getCells()])
    stat.value = s
    cells.value = c
  } catch (e) {
    console.error('铝厂概览加载失败', e)
  }
})
</script>

<style scoped lang="scss">
.al-status-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;

  .al-status-label {
    width: 70px;
    flex-shrink: 0;
    text-align: right;
    color: #8d9db8;
    font-size: 12px;
  }

  :deep(.el-progress) {
    flex: 1;
  }

  .al-status-count {
    width: 60px;
    text-align: right;
    font-size: 12px;
    font-weight: 600;
  }
}
</style>
