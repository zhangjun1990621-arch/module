import type { Component } from 'vue'

/**
 * 自定义组件注册表
 * ------------------------------------------------
 * 当平台页面配置为 type: "custom" 时，DynamicPage 会根据 component 字段
 * 从此注册表中查找对应的 Vue 组件并渲染。
 *
 * 新增平台的自定义组件时：
 * 1. 将组件文件放到 src/views/<platform>/ 目录下
 * 2. 在此注册表中添加映射条目
 * 3. 在数据库平台配置中使用 "type": "custom", "component": "<name>"
 */
const registry: Record<string, () => Promise<{ default: Component }>> = {
  // === 铝厂平台自定义组件 ===
  AluminumOverview: () => import('@/views/aluminum/OverviewView.vue'),
  AluminumCellMonitor: () => import('@/views/aluminum/CellMonitorView.vue'),
  AluminumPointRealtime: () => import('@/views/aluminum/PointRealtimeView.vue'),
  AluminumAlarm: () => import('@/views/aluminum/AlarmView.vue'),
  AluminumHistory: () => import('@/views/aluminum/HistoryView.vue'),

  // === 光伏平台自定义组件 ===
  PvDashboard: () => import('@/views/dashboard/DashboardView.vue'),
  PvDeviceDetail: () => import('@/views/device/DeviceDetailView.vue'),
  PvEventList: () => import('@/views/event/EventListView.vue'),
  PvOTA: () => import('@/views/ota/OTAView.vue'),
  PvStation: () => import('@/views/station/StationView.vue'),
}

/** 根据组件名称获取异步加载函数 */
export function getCustomComponent(
  name: string
): (() => Promise<{ default: Component }>) | undefined {
  return registry[name]
}
