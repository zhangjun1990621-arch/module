<template>
  <div class="dynamic-chart">
    <div class="al-card chart-card" v-loading="loading">
      <div class="chart-header">
        <span class="chart-title">{{ title }}</span>
        <el-button text type="primary" size="small" @click="fetchData">刷新</el-button>
      </div>
      <div ref="chartRef" class="chart-container"></div>
      <el-empty v-if="!loading && !hasData" description="暂无数据" class="chart-empty" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import * as echarts from 'echarts'
import request from '@/api/request'

const props = defineProps<{
  api: string
  chartType: string
  platform?: string
  title?: string
}>()

const loading = ref(false)
const chartRef = ref<HTMLElement>()
const hasData = ref(false)
let chartInstance: echarts.ECharts | null = null

/** 获取图表数据 */
async function fetchData() {
  if (!props.api) return
  loading.value = true
  try {
    // 构建请求路径：如果 api 以 / 开头则直接使用，否则拼接 platform 前缀
    let url = props.api
    if (!url.startsWith('/') && props.platform) {
      url = `/${props.platform}/${url}`
    }
    const res = await request.get(url)
    const rawData = res.data
    renderChart(rawData)
  } catch {
    hasData.value = false
  } finally {
    loading.value = false
  }
}

/** 渲染图表 */
function renderChart(data: any) {
  if (!chartRef.value || !chartInstance) return

  let option: echarts.EChartsOption

  const type = props.chartType || 'line'

  if (type === 'pie') {
    // 饼图数据格式：[{ name, value }] 或 { categories, values }
    let pieData = data
    if (data && data.categories && data.values) {
      pieData = data.categories.map((cat: string, i: number) => ({
        name: cat,
        value: data.values[i]
      }))
    } else if (data && data.list) {
      pieData = data.list
    }

    if (!Array.isArray(pieData) || pieData.length === 0) {
      hasData.value = false
      chartInstance.clear()
      return
    }

    hasData.value = true
    option = {
      tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
      legend: { bottom: 10, textStyle: { color: '#8b949e' } },
      series: [
        {
          type: 'pie',
          radius: ['40%', '70%'],
          center: ['50%', '45%'],
          itemStyle: {
            borderColor: '#161b22',
            borderWidth: 2
          },
          label: { color: '#8b949e' },
          data: pieData
        }
      ],
      color: ['#4b3fe3', '#58a6ff', '#3fb950', '#f0883e', '#f85149', '#d29922', '#a371f7']
    }
  } else {
    // 折线图 / 柱状图数据格式：{ categories, series: [{ name, data }] }
    // 或 { xaxis, series } 或直接 { categories, values }
    const categories = data?.categories || data?.xaxis || data?.labels || []
    let series = data?.series || []

    // 兼容单系列数据 { categories, values }
    if (series.length === 0 && data?.values) {
      series = [{ name: props.title || '数据', data: data.values }]
    }

    if (!categories.length || !series.length) {
      hasData.value = false
      chartInstance.clear()
      return
    }

    hasData.value = true
    option = {
      tooltip: { trigger: 'axis' },
      legend: {
        bottom: 0,
        textStyle: { color: '#8b949e' },
        data: series.map((s: any) => s.name)
      },
      grid: { top: 30, right: 20, bottom: 50, left: 50 },
      xAxis: {
        type: 'category',
        data: categories,
        axisLabel: { color: '#6e7681' },
        axisLine: { lineStyle: { color: '#30363d' } }
      },
      yAxis: {
        type: 'value',
        axisLabel: { color: '#6e7681' },
        axisLine: { lineStyle: { color: '#30363d' } },
        splitLine: { lineStyle: { color: '#21262d' } }
      },
      series: series.map((s: any) => ({
        name: s.name,
        type: type === 'bar' ? 'bar' : 'line',
        data: s.data,
        smooth: type === 'line',
        areaStyle: type === 'line' ? { opacity: 0.15 } : undefined,
        itemStyle: { borderRadius: type === 'bar' ? [4, 4, 0, 0] : undefined }
      })),
      color: ['#4b3fe3', '#58a6ff', '#3fb950', '#f0883e', '#f85149', '#d29922']
    }
  }

  chartInstance.setOption(option, true)
}

/** 初始化图表 */
function initChart() {
  if (!chartRef.value) return
  chartInstance = echarts.init(chartRef.value)
  fetchData()
}

/** 窗口大小变化时重绘 */
function handleResize() {
  chartInstance?.resize()
}

onMounted(async () => {
  await nextTick()
  initChart()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  chartInstance?.dispose()
  chartInstance = null
})

watch(
  () => [props.api, props.chartType],
  () => {
    fetchData()
  }
)
</script>

<style scoped lang="scss">
.dynamic-chart {
  .chart-card {
    padding: 0;
    overflow: hidden;

    .chart-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 16px 20px;
      border-bottom: 1px solid #21262d;

      .chart-title {
        font-size: 15px;
        font-weight: 600;
        color: #e6edf3;
      }
    }

    .chart-container {
      width: 100%;
      height: 400px;
      padding: 12px;
    }

    .chart-empty {
      padding: 80px 0;
    }
  }
}
</style>
