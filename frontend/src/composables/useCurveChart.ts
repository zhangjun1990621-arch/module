import { ref, shallowRef, onMounted, onBeforeUnmount, watch, type Ref } from 'vue'
import * as echarts from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent, DataZoomComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([LineChart, GridComponent, TooltipComponent, LegendComponent, DataZoomComponent, CanvasRenderer])

export interface CurvePoint {
  time: string
  values: number[]
}

export function useCurveChart(
  chartRef: Ref<HTMLElement | null>,
  labels: string[],
  colors: string[],
  yUnit: string,
  maxPoints = 60
) {
  const chart = shallowRef<echarts.ECharts | null>(null)
  const seriesData = labels.map(() => ref<number[]>([]))
  const timeData = ref<string[]>([])

  const getOption = (): echarts.EChartsCoreOption => ({
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(24,36,52,0.95)',
      borderColor: '#253650',
      textStyle: { color: '#e6ecf5', fontSize: 11 },
      formatter: (params: any) => {
        if (!Array.isArray(params)) return ''
        let tip = `<div style="font-size:11px">${params[0].axisValue}</div>`
        params.forEach((p: any) => {
          tip += `<div style="display:flex;align-items:center;gap:4px;margin-top:2px">
            <span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:${p.color}"></span>
            ${p.seriesName}: <b>${p.value}</b> ${yUnit}
          </div>`
        })
        return tip
      }
    },
    legend: {
      data: labels,
      top: 4,
      right: 10,
      textStyle: { color: '#566880', fontSize: 10 },
      itemWidth: 14,
      itemHeight: 8
    },
    grid: {
      left: 48,
      right: 12,
      top: 32,
      bottom: 48
    },
    xAxis: {
      type: 'category',
      data: timeData.value,
      axisLine: { lineStyle: { color: '#253650' } },
      axisLabel: { color: '#566880', fontSize: 10, rotate: 0 },
      axisTick: { show: false }
    },
    yAxis: {
      type: 'value',
      name: yUnit,
      nameTextStyle: { color: '#566880', fontSize: 10, padding: [0, 0, 0, -10] },
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: { lineStyle: { color: '#253650', type: 'dashed' } },
      axisLabel: { color: '#566880', fontSize: 10 }
    },
    dataZoom: [{
      type: 'inside',
      start: 0,
      end: 100
    }],
    series: labels.map((name, i) => ({
      name,
      type: 'line' as const,
      data: seriesData[i].value,
      smooth: true,
      symbol: 'none',
      lineStyle: { width: 2, color: colors[i] },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: colors[i] + '30' },
          { offset: 1, color: colors[i] + '05' }
        ])
      }
    })),
    animation: false
  })

  function pushPoint(values: number[]) {
    const now = new Date()
    const t = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}`
    timeData.value.push(t)
    if (timeData.value.length > maxPoints) timeData.value.shift()
    values.forEach((v, i) => {
      seriesData[i].value.push(v)
      if (seriesData[i].value.length > maxPoints) seriesData[i].value.shift()
    })
    if (chart.value) {
      chart.value.setOption(getOption())
    }
  }

  function initChart() {
    if (chartRef.value && !chart.value) {
      chart.value = echarts.init(chartRef.value, undefined, { renderer: 'canvas' })
      chart.value.setOption(getOption())
    }
  }

  function disposeChart() {
    chart.value?.dispose()
    chart.value = null
  }

  return { pushPoint, initChart, disposeChart }
}
