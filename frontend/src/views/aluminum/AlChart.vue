<template>
  <div ref="chartRef" class="al-chart-box" :style="{ height: height }" />
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import * as echarts from 'echarts'

const props = defineProps<{
  option: echarts.EChartsOption
  height?: string
}>()

const chartRef = ref<HTMLElement | null>(null)
let chart: echarts.ECharts | null = null

const height = props.height || '280px'

function initChart() {
  if (!chartRef.value) return
  chart = echarts.init(chartRef.value, 'dark')
  chart.setOption(props.option)
}

function resize() {
  chart?.resize()
}

watch(
  () => props.option,
  (val) => {
    if (chart) {
      chart.setOption(val, true)
    } else {
      nextTick(initChart)
    }
  },
  { deep: true }
)

onMounted(() => {
  initChart()
  window.addEventListener('resize', resize)
})

onUnmounted(() => {
  window.removeEventListener('resize', resize)
  chart?.dispose()
  chart = null
})
</script>
